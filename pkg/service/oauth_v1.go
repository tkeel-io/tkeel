package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tkeel-io/tkeel/pkg/client/kubernetes"
	t_model "github.com/tkeel-io/tkeel/pkg/model"

	"github.com/casbin/casbin/v2"
	"github.com/coreos/go-oidc"
	dapr "github.com/dapr/go-sdk/client"
	oauth2v4 "github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	kitErr "github.com/tkeel-io/kit/errors"
	"github.com/tkeel-io/kit/log"
	transportHTTP "github.com/tkeel-io/kit/transport/http"
	"github.com/tkeel-io/security/authn/idprovider"
	oidcprovider "github.com/tkeel-io/security/authn/idprovider/oidc"
	"github.com/tkeel-io/security/model"
	pb "github.com/tkeel-io/tkeel/api/security_oauth/v1"
	"github.com/tkeel-io/tkeel/pkg/util"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

const (
	DefaultClient         = "tkeel"
	DefaultClientSecurity = "tkeel"
	DefaultClientDomain   = "tkeel.io"
	TokenTypeBearer       = "Bearer"
)

var DefaultGrantType = []oauth2v4.GrantType{oauth2v4.AuthorizationCode, oauth2v4.Implicit, oauth2v4.PasswordCredentials, oauth2v4.Refreshing}

type OauthService struct {
	Config     *TokenConf
	Manager    *manage.Manager
	UserDB     *gorm.DB
	DaprStore  string
	DaprClient dapr.Client
	k8s        *kubernetes.Client
	RBACOp     *casbin.SyncedEnforcer
	pb.UnimplementedOauthServer
}

type TokenConf struct {
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration

	TokenType         string               // token type.
	AllowedGrantTypes []oauth2v4.GrantType // allow the grant type.
}

// AuthorizeRequest authorization request.
type AuthorizeRequest struct {
	ResponseType        oauth2v4.ResponseType
	ClientID            string
	Scope               string
	RedirectURI         string
	State               string
	UserID              string
	CodeChallenge       string
	CodeChallengeMethod oauth2v4.CodeChallengeMethod
	AccessTokenExp      time.Duration
	Request             *http.Request
}

func NewOauthService(m *manage.Manager, userDB *gorm.DB, conf *TokenConf, daprClient dapr.Client, daprstore string, k8s *kubernetes.Client, rbac *casbin.SyncedEnforcer) *OauthService {
	if userDB == nil {
		log.Error("nil db")
		panic("nil db")
	}
	oauthServer := &OauthService{UserDB: userDB, Config: conf, Manager: m, DaprClient: daprClient, DaprStore: daprstore, k8s: k8s, RBACOp: rbac}
	return oauthServer
}

func (s *OauthService) Authorize(ctx context.Context, req *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	authorizeReq := &AuthorizeRequest{
		ResponseType: oauth2v4.ResponseType(req.GetResponseType()),
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
	provider, err := idprovider.GetIdentityProvider(req.GetTenantId())
	if err != nil {
		item, _ := s.DaprClient.GetState(ctx, s.DaprStore, KeyOfTenantIdentityProvider(req.GetTenantId()))
		if item != nil {
			ProviderRegister(ctx, item.Value)
		}
		provider, err = idprovider.GetIdentityProvider(req.GetTenantId())
		if err != nil {
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
	}

	switch provider.Type() {
	case "OIDCIdentityProvider":
		if req.GetCode() != "" {
			identity, err := provider.AuthenticateCode(req.GetCode())
			if err != nil {
				log.Error(err)
				return nil, pb.OauthErrUnknown()
			}
			userDao := &model.User{}
			whereDao := model.User{ExternalID: identity.GetExternalID(), TenantID: req.GetTenantId()}
			assignDao := model.User{UserName: identity.GetUsername(), Email: identity.GetEmail()}
			err = userDao.FirstOrAssignCreate(s.UserDB, whereDao, assignDao)
			if err != nil {
				log.Error(err)
				return nil, pb.OauthErrServerError()
			}
			roleDao := &model.Role{}
			count, roles, _ := roleDao.List(s.UserDB, map[string]interface{}{"tenant_id": req.GetTenantId(), "name": t_model.TkeelTenantAdminRole}, nil, "")
			if count < 1 {
				log.Errorf("%s tenant hasn't admin", req.GetTenantId())
				return nil, pb.OauthErrServerError()
			}
			adminRBACs := s.RBACOp.GetFilteredGroupingPolicy(1, roles[0].ID, req.GetTenantId())
			if len(adminRBACs) == 0 {
				adminRBAC := []string{userDao.ID, roles[0].ID, req.GetTenantId()}
				s.RBACOp.AddGroupingPolicy(adminRBAC)
			}
			tgr := &oauth2v4.TokenGenerateRequest{ClientID: DefaultClient, ClientSecret: DefaultClientSecurity, UserID: userDao.ID}
			ti, err := s.Manager.GenerateAccessToken(ctx, oauth2v4.PasswordCredentials, tgr)
			if err != nil {
				log.Error(err)
				return nil, pb.OauthErrServerError()
			}

			conf, err := s.k8s.GetDeploymentConfig(ctx)
			if err != nil {
				log.Errorf("error get deployment config: %s", err)
				return nil, pb.OauthErrServerError()
			}
			if req.DisableRedirect {
				return &pb.TokenResponse{AccessToken: ti.GetAccess(),
					RefreshToken: ti.GetRefresh(),
					ExpiresIn:    int64(ti.GetAccessExpiresIn() / time.Second),
					TokenType:    s.Config.TokenType}, nil
			}
			redirect := fmt.Sprintf("http://%s:%s/auth/redirect?token_type=%s&access_token=%s&refresh_token=%s&expires_in=%d",
				conf.Host.Tenant, conf.Port, s.Config.TokenType, ti.GetAccess(), ti.GetRefresh(), int64(ti.GetAccessExpiresIn()/time.Second))
			return nil, kitErr.NewRedirect(redirect)
		}
		return &pb.TokenResponse{RedirectUrl: provider.AuthCodeURL("", "")}, nil
	}

	return nil, pb.OauthErrInvalidRequest()
}

func (s *OauthService) Authenticate(ctx context.Context, empty *emptypb.Empty) (*pb.AuthenticateResponse, error) {
	header := transportHTTP.HeaderFromContext(ctx)
	accessToken, ok := bearerAuth(header)
	if !ok {
		return nil, pb.OauthErrInvalidAccessToken()
	}
	token, err := s.Manager.LoadAccessToken(ctx, accessToken)
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrInvalidAccessToken()
	}
	userID := token.GetUserID()
	user := &model.User{}
	_, users, err := user.QueryByCondition(s.UserDB, map[string]interface{}{"id": userID}, nil, "")
	if err != nil || len(users) != 1 {
		log.Error(err)
		s.Manager.RemoveAccessToken(ctx, accessToken)
		return nil, pb.OauthErrInvalidAccessToken()
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
func (s *OauthService) GetAuthorizeToken(ctx context.Context, req *AuthorizeRequest) (oauth2v4.TokenInfo, error) {
	// check the client allows the grant type.
	tgr := &oauth2v4.TokenGenerateRequest{
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
func (s *OauthService) ValidationTokenRequest(r *pb.TokenRequest) (oauth2v4.GrantType, *oauth2v4.TokenGenerateRequest, error) {
	gt := oauth2v4.GrantType(r.GetGrantType())
	if gt.String() == "" {
		return "", nil, pb.OauthErrUnsupportedGrantType()
	}

	tgr := &oauth2v4.TokenGenerateRequest{
		ClientID:     DefaultClient,
		ClientSecret: DefaultClientSecurity,
	}

	switch gt {
	case oauth2v4.AuthorizationCode:
		tgr.RedirectURI = r.GetRedirectUri()
		tgr.Code = r.GetCode()
		if tgr.RedirectURI == "" ||
			tgr.Code == "" {
			return "", nil, pb.OauthErrInvalidRequest()
		}

	case oauth2v4.PasswordCredentials:
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
	case oauth2v4.Refreshing:
		tgr.Refresh = r.GetRefreshToken()
		if tgr.Refresh == "" {
			return "", nil, pb.OauthErrInvalidRequest()
		}
	}
	return gt, tgr, nil
}

// CheckGrantType check allows grant type.
func (s *OauthService) CheckGrantType(gt oauth2v4.GrantType) bool {
	for _, agt := range s.Config.AllowedGrantTypes {
		if agt == gt {
			return true
		}
	}
	return false
}

// GetAccessToken access token. //nolint.
func (s *OauthService) GetAccessToken(ctx context.Context, gt oauth2v4.GrantType, tgr *oauth2v4.TokenGenerateRequest) (oauth2v4.TokenInfo,
	error) {
	if allowed := s.CheckGrantType(gt); !allowed {
		return nil, pb.OauthErrUnauthorizedClient()
	}
	switch gt {
	case oauth2v4.AuthorizationCode:
		ti, err := s.Manager.GenerateAccessToken(ctx, gt, tgr)
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf("generete token %w", err)
		}
		return ti, nil
	case oauth2v4.PasswordCredentials, oauth2v4.ClientCredentials:
		info, err := s.Manager.GenerateAccessToken(ctx, gt, tgr)
		if err != nil {
			return info, fmt.Errorf("%w", err)
		}
		return info, nil
	case oauth2v4.Refreshing:
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
func bearerAuth(header http.Header) (string, bool) {
	auth := header.Get("Authorization")
	prefix := "Bearer "
	token := ""
	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	}
	return token, token != ""
}

func (s *OauthService) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	uDao := &model.User{}
	total, users, err := uDao.QueryByCondition(s.UserDB, map[string]interface{}{"password": req.GetBody().GetResetKey()}, nil, "")
	if err != nil || total != 1 {
		log.Error(err)
		return nil, pb.OauthErrInvalidResetPwd()
	}
	user := &model.User{Password: req.GetBody().GetNewPassword()}
	user.Encrypt()
	updates := map[string]interface{}{"password": user.Password}
	err = user.Update(s.UserDB, users[0].TenantID, users[0].ID, updates)
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrServerError()
	}
	return &pb.ResetPasswordResponse{TenantId: users[0].TenantID, HasReset: true, Username: users[0].UserName}, nil
}

func (s *OauthService) OIDCRegister(ctx context.Context, req *pb.OIDCRegisterRequest) (*pb.OIDCRegisterResponse, error) {
	if req.GetBody().GetTenantId() == "" || req.GetBody().GetClientId() == "" || req.GetBody().GetClientSecret() == "" || req.GetBody().GetRedirectUrl() == "" {
		log.Error("invalid oidc register params")
		return nil, pb.OauthErrInvalidRequest()
	}
	var err error
	oidcProvider := &oidcprovider.OIDCProvider{Issuer: req.GetBody().GetIssuer(), ClientID: req.GetBody().GetClientId(),
		ClientSecret: req.GetBody().GetClientSecret()}
	bytesBody, err := json.Marshal(req.GetBody())
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrInvalidRequest()
	}
	json.Unmarshal(bytesBody, oidcProvider)
	oauth2Endpoint := oauth2.Endpoint{}
	if req.GetBody().GetIssuer() != "" {
		oidcProvider.Provider, err = oidc.NewProvider(ctx, req.GetBody().GetIssuer())
		if err != nil {
			log.Error(err)
			return nil, pb.OauthErrUnknown()
		}
		oauth2Endpoint = oidcProvider.Provider.Endpoint()
	} else {
		oauth2Endpoint.AuthURL = oidcProvider.Endpoint.AuthURL
		oauth2Endpoint.TokenURL = oidcProvider.Endpoint.TokenURL
	}
	oauth2Config := &oauth2.Config{
		ClientID:     oidcProvider.ClientID,
		ClientSecret: oidcProvider.ClientSecret,
		RedirectURL:  oidcProvider.RedirectURL,
		Scopes:       oidcProvider.Scopes,
		Endpoint:     oauth2Endpoint,
	}
	oidcProvider.OAuth2Config = oauth2Config
	idprovider.RegisterIdentityProvider(req.GetBody().GetTenantId(), oidcProvider)
	identityInfo := map[string]interface{}{"type": "OIDCIdentityProvider", "tenant_id": req.GetBody().GetTenantId(), "info": req.GetBody()}
	bytesInfo, _ := json.Marshal(identityInfo)
	s.DaprClient.SaveState(ctx, s.DaprStore, KeyOfTenantIdentityProvider(req.GetBody().GetTenantId()), bytesInfo)
	return &pb.OIDCRegisterResponse{Ok: true}, nil
}
func (s *OauthService) TokenRevoke(ctx context.Context, req *pb.TokenRevokeRequest) (*pb.TokenRevokeResponse, error) {
	ti, err := s.Manager.LoadRefreshToken(ctx, req.GetBody().GetRefreshToken())
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrInvalidAccessToken()
	}
	userDao := &model.User{}
	num, users, err := userDao.QueryByCondition(s.UserDB, map[string]interface{}{"id": ti.GetUserID()}, nil, "")
	if err != nil || num != 1 {
		log.Error(err)
		return nil, pb.OauthErrInvalidRequest()
	}
	s.Manager.RemoveRefreshToken(ctx, ti.GetRefresh())
	s.Manager.RemoveAccessToken(ctx, ti.GetAccess())
	return &pb.TokenRevokeResponse{
		Revoked:  true,
		TenantId: users[0].TenantID,
	}, nil
}

func (s *OauthService) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	user, err := util.GetUser(ctx)
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrInvalidAccessToken()
	}
	useraDao := &model.User{Password: req.GetBody().GetNewPassword()}
	useraDao.Encrypt()
	err = useraDao.Update(s.UserDB, user.Tenant, user.User, map[string]interface{}{"password": useraDao.Password})
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrServerError()
	}
	ti, err := s.Manager.LoadRefreshToken(ctx, req.GetBody().GetRefreshToken())
	if err != nil {
		log.Error(err)
		return nil, pb.OauthErrInvalidAccessToken()
	}
	s.Manager.RemoveRefreshToken(ctx, ti.GetRefresh())
	s.Manager.RemoveAccessToken(ctx, ti.GetAccess())

	return &pb.UpdatePasswordResponse{TenantId: user.Tenant}, nil
}

func KeyOfTenantIdentityProvider(tenantID string) string {
	return fmt.Sprintf("tkeel:provider:%s", tenantID)
}

func ProviderRegister(ctx context.Context, data []byte) error {
	dataMap := map[string]interface{}{}
	err := json.Unmarshal(data, &dataMap)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("%w", err)
	}
	tenantID := dataMap["tenant_id"]
	providerType, ok := dataMap["type"]
	if !ok {
		return errors.New("invalid type")
	}
	switch providerType.(string) {
	case "OIDCIdentityProvider":
		infoBytes, err := json.Marshal(dataMap["info"])
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		oidcProvider := &oidcprovider.OIDCProvider{}
		infoMap := map[string]interface{}{}
		json.Unmarshal(infoBytes, oidcProvider)
		json.Unmarshal(infoBytes, &infoMap)
		clientSecret, ok := infoMap["client_secret"].(string)
		if !ok {
			return errors.New("not client secret")
		}
		oidcProvider.ClientSecret = clientSecret
		oauth2Endpoint := oauth2.Endpoint{}
		if oidcProvider.Issuer != "" {
			oidcProvider.Provider, err = oidc.NewProvider(ctx, oidcProvider.Issuer)
			if err != nil {
				log.Error(err)
				return fmt.Errorf("%w", err)
			}
			oauth2Endpoint = oidcProvider.Provider.Endpoint()
		} else {
			oauth2Endpoint.AuthURL = oidcProvider.Endpoint.AuthURL
			oauth2Endpoint.TokenURL = oidcProvider.Endpoint.TokenURL
		}
		oauth2Config := &oauth2.Config{
			ClientID:     oidcProvider.ClientID,
			ClientSecret: oidcProvider.ClientSecret,
			RedirectURL:  oidcProvider.RedirectURL,
			Scopes:       oidcProvider.Scopes,
			Endpoint:     oauth2Endpoint,
		}
		oidcProvider.OAuth2Config = oauth2Config
		idprovider.RegisterIdentityProvider(tenantID.(string), oidcProvider)
	default:
		return errors.New("invalid provider type")
	}
	return nil
}
