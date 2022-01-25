package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	pb "github.com/tkeel-io/tkeel/api/security_oauth/v1"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/tkeel-io/kit/log"
	transportHTTP "github.com/tkeel-io/kit/transport/http"
	"github.com/tkeel-io/security/model"
	"gorm.io/gorm"
)

const (
	DefaultClient         = "tkeel"
	DefaultClientSecurity = "tkeel"
	DefaultClientDomain   = "tkeel.io"
	TokenTypeBearer       = "Bearer"
)

var DefaultGrantType = []oauth2.GrantType{oauth2.AuthorizationCode, oauth2.Implicit, oauth2.PasswordCredentials, oauth2.Refreshing}

type OauthService struct {
	Config  *TokenConf
	Manager *manage.Manager
	UserDB  *gorm.DB
	pb.UnimplementedOauthServer
}

type TokenConf struct {
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration

	TokenType         string             // token type.
	AllowedGrantTypes []oauth2.GrantType // allow the grant type.
}

// AuthorizeRequest authorization request.
type AuthorizeRequest struct {
	ResponseType        oauth2.ResponseType
	ClientID            string
	Scope               string
	RedirectURI         string
	State               string
	UserID              string
	CodeChallenge       string
	CodeChallengeMethod oauth2.CodeChallengeMethod
	AccessTokenExp      time.Duration
	Request             *http.Request
}

func NewOauthService(userDB *gorm.DB, conf *TokenConf, tokenStore oauth2.TokenStore, generator oauth2.AccessGenerate, client oauth2.ClientInfo) *OauthService {
	manager := manage.NewDefaultManager()
	tokenConf := manage.DefaultAuthorizeCodeTokenCfg
	if conf.AccessTokenExp != 0 && conf.RefreshTokenExp != 0 {
		tokenConf.AccessTokenExp = conf.AccessTokenExp
		tokenConf.RefreshTokenExp = conf.RefreshTokenExp
	}

	log.Info(tokenConf)
	clientStore := store.NewClientStore()
	if client == nil {
		client = &models.Client{ID: DefaultClient, Secret: DefaultClientSecurity, Domain: DefaultClientDomain}
	}
	clientStore.Set(client.GetID(), client)
	if tokenStore == nil {
		tokenStore, _ = store.NewMemoryTokenStore()
	}
	if generator == nil {
		generator = generates.NewAccessGenerate()
	}

	manager.SetPasswordTokenCfg(tokenConf)
	manager.MapClientStorage(clientStore)
	manager.MapTokenStorage(tokenStore)
	manager.MapAccessGenerate(generator)
	if userDB == nil {
		log.Error("nil db")
		panic("nil db")
	}
	oauthServer := &OauthService{UserDB: userDB, Config: conf, Manager: manager}
	return oauthServer
}

func (s *OauthService) Authorize(ctx context.Context, req *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	authorizeReq := &AuthorizeRequest{
		ResponseType: oauth2.ResponseType(req.GetResponseType()),
		RedirectURI:  req.GetRedirectUri(),
		ClientID:     DefaultClient,
		State:        req.GetState(),
	}
	// user authorization.
	user, err := model.AuthenticateUser(s.UserDB, req.GetTenantId(), req.GetUsername(), req.GetPassword())
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrInvalidRequest()
	}
	authorizeReq.UserID = user.ID
	ti, err := s.GetAuthorizeToken(ctx, authorizeReq)
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrUnauthorizedClient()
	}
	return &pb.AuthorizeResponse{Code: ti.GetCode()}, nil
}

func (s *OauthService) Token(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	gt, tgr, err := s.ValidationTokenRequest(req)
	if err != nil {
		return nil, err
	}
	ti, err := s.GetAccessToken(ctx, gt, tgr)
	if err != nil {
		return nil, pb.OauthErrServerError()
	}
	log.Info(ti.GetAccessExpiresIn())
	log.Info(ti.GetAccessExpiresIn() / time.Second)

	return &pb.TokenResponse{
		AccessToken:  ti.GetAccess(),
		RefreshToken: ti.GetRefresh(),
		ExpiresIn:    int64(ti.GetAccessExpiresIn() / time.Second),
		TokenType:    s.Config.TokenType,
	}, nil
}

func (s *OauthService) Authenticate(ctx context.Context, empty *emptypb.Empty) (*pb.AuthenticateResponse, error) {
	header := transportHTTP.HeaderFromContext(ctx)
	accessToken, ok := s.bearerAuth(&header)
	if !ok {
		return nil, pb.OauthErrInvalidAccessToken()
	}
	token, err := s.Manager.LoadAccessToken(ctx, accessToken)
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrServerError()
	}
	userID := token.GetUserID()
	user := &model.User{}
	_, users, err := user.QueryByCondition(s.UserDB, map[string]interface{}{"id": userID}, nil)
	if err != nil || len(users) != 1 {
		log.Error(err)
		return nil, pb.OauthErrServerError()
	}

	return &pb.AuthenticateResponse{
		ExpiresIn:  int64(token.GetAccessExpiresIn() / time.Second),
		Username:   users[0].UserName,
		UserId:     users[0].ID,
		ExternalId: users[0].ExternalID,
		NickName:   users[0].NickName,
		Avatar:     users[0].Avatar,
		TenantId:   users[0].TenantID,
	}, nil
}

// GetAuthorizeToken get authorization token(code). //nolint.
func (s *OauthService) GetAuthorizeToken(ctx context.Context, req *AuthorizeRequest) (oauth2.TokenInfo, error) {
	// check the client allows the grant type.
	tgr := &oauth2.TokenGenerateRequest{
		ClientID:       req.ClientID,
		UserID:         req.UserID,
		RedirectURI:    req.RedirectURI,
		Scope:          req.Scope,
		AccessTokenExp: req.AccessTokenExp,
		Request:        req.Request,
	}

	tgr.CodeChallenge = req.CodeChallenge
	tgr.CodeChallengeMethod = req.CodeChallengeMethod

	info, err := s.Manager.GenerateAuthToken(ctx, req.ResponseType, tgr)
	if err != nil {
		return info, fmt.Errorf("generate %w", err)
	}
	return info, nil
}

// ValidationTokenRequest the token request validation.
func (s *OauthService) ValidationTokenRequest(r *pb.TokenRequest) (oauth2.GrantType, *oauth2.TokenGenerateRequest, error) {
	gt := oauth2.GrantType(r.GetGrantType())
	if gt.String() == "" {
		return "", nil, pb.OauthErrUnsupportedGrantType()
	}

	tgr := &oauth2.TokenGenerateRequest{
		ClientID:     DefaultClient,
		ClientSecret: DefaultClientSecurity,
	}

	switch gt {
	case oauth2.AuthorizationCode:
		tgr.RedirectURI = r.GetRedirectUri()
		tgr.Code = r.GetCode()
		if tgr.RedirectURI == "" ||
			tgr.Code == "" {
			return "", nil, pb.OauthErrInvalidRequest()
		}

	case oauth2.PasswordCredentials:
		username, password := r.GetUsername(), r.GetPassword()
		if username == "" || password == "" {
			return "", nil, pb.OauthErrInvalidRequest()
		}
		user, err := model.AuthenticateUser(s.UserDB, r.GetTenantId(),
			r.GetUsername(), r.GetPassword())
		if err != nil {
			return "", nil, pb.OauthErrInvalidRequest()
		}
		tgr.UserID = user.ID
	case oauth2.Refreshing:
		tgr.Refresh = r.GetRefreshToken()
		if tgr.Refresh == "" {
			return "", nil, pb.OauthErrInvalidRequest()
		}
	}
	return gt, tgr, nil
}

// CheckGrantType check allows grant type.
func (s *OauthService) CheckGrantType(gt oauth2.GrantType) bool {
	for _, agt := range s.Config.AllowedGrantTypes {
		if agt == gt {
			return true
		}
	}
	return false
}

// GetAccessToken access token. //nolint.
func (s *OauthService) GetAccessToken(ctx context.Context, gt oauth2.GrantType, tgr *oauth2.TokenGenerateRequest) (oauth2.TokenInfo,
	error) {
	if allowed := s.CheckGrantType(gt); !allowed {
		return nil, pb.OauthErrUnauthorizedClient()
	}
	switch gt {
	case oauth2.AuthorizationCode:
		ti, err := s.Manager.GenerateAccessToken(ctx, gt, tgr)
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf("generete token %w", err)
		}
		return ti, nil
	case oauth2.PasswordCredentials, oauth2.ClientCredentials:
		info, err := s.Manager.GenerateAccessToken(ctx, gt, tgr)
		if err != nil {
			return info, fmt.Errorf("%w", err)
		}
		return info, nil
	case oauth2.Refreshing:
		ti, err := s.Manager.RefreshAccessToken(ctx, tgr)
		if err != nil {
			log.Error(err)
			return nil, pb.OauthErrInvalidRequest()
		}
		return ti, nil
	}
	return nil, pb.OauthErrUnsupportedGrantType()
}

// bearerAuth parse bearer token.
func (s *OauthService) bearerAuth(header *http.Header) (string, bool) {
	auth := header.Get("Authorization")
	prefix := "Bearer "
	token := ""
	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	}
	return token, token != ""
}
