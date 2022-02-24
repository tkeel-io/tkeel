package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/tkeel-io/kit/log"
	transport_http "github.com/tkeel-io/kit/transport/http"
	pb "github.com/tkeel-io/tkeel/api/oauth2/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/token"
	"github.com/tkeel-io/tkeel/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	expires = 1 * time.Hour
)

var ErrSecretNotMatch = errors.New("secret not match")

type Oauth2ServiceV1 struct {
	pb.UnimplementedOauth2Server

	kvOp              kv.Operator
	pluginIDWhiteList []string
	pluginOp          plugin.Operator
	secretProvider    token.Provider
}

func NewOauth2ServiceV1(adminPasswd string, kvOp kv.Operator, pOp plugin.Operator) *Oauth2ServiceV1 {
	values, ver, err := kvOp.Get(context.TODO(), model.KeyAdminPassword)
	if err != nil {
		log.Fatalf("error init rudder admin password: %s", err)
		return nil
	}
	if ver == "" {
		if err = kvOp.Create(context.TODO(), model.KeyAdminPassword,
			[]byte(base64.StdEncoding.EncodeToString([]byte(adminPasswd)))); err != nil {
			log.Fatalf("error create rudder admin password: %s", err)
			return nil
		}
	}
	log.Debugf("rudder admin password: %s", string(values))
	secret := util.RandStringBytesMaskImpr(0, 10)
	return &Oauth2ServiceV1{
		kvOp:              kvOp,
		pluginIDWhiteList: []string{"rudder", "keel", "core"},
		pluginOp:          pOp,
		secretProvider:    token.InitProvider(secret, "", ""),
	}
}

func (s *Oauth2ServiceV1) AddPluginWhiteList(ctx context.Context,
	req *pb.AddPluginWhiteListRequest) (*emptypb.Empty, error) {
	if s.checkPluginWhiteList(req.PluginId) {
		return nil, fmt.Errorf("error duplicate add")
	}
	s.pluginIDWhiteList = append(s.pluginIDWhiteList, req.PluginId)
	return &emptypb.Empty{}, nil
}

func (s *Oauth2ServiceV1) IssuePluginToken(ctx context.Context, req *pb.IssuePluginTokenRequest) (*pb.IssueTokenResponse, error) {
	pluginID := req.ClientId
	if pluginID == "" {
		log.Errorf("error invalid plugin id: empty string")
		return nil, pb.Oauth2ErrInvalidPluginId()
	}
	if !s.checkPluginWhiteList(pluginID) {
		plugin, err := s.pluginOp.Get(ctx, pluginID)
		if err != nil {
			log.Errorf("error issue(%s) oauth2 token: %s", pluginID, err)
			return nil, pb.Oauth2ErrInternalStore()
		}
		if err = s.checkPluginSecret(plugin.Secret, req.ClientSecret); err != nil {
			log.Errorf("error issue(%s) oauth2 token(%s -- %s): %s", pluginID, plugin.Secret, req.ClientSecret, err)
			return nil, pb.Oauth2ErrSecretNotMatch()
		}
	}
	token, _, err := s.genToken("plugin", "plugin_id", pluginID)
	if err != nil {
		log.Errorf("error issue(%s) oauth2 token gen plugin token: %s", pluginID, err)
		return nil, pb.Oauth2ErrUnknown()
	}
	log.Debugf("issue(%s) oauth2 token: %s", pluginID, token)
	return &pb.IssueTokenResponse{
		TokenType:   "Bearer",
		AccessToken: token,
		ExpiresIn:   int32(expires.Seconds()),
	}, nil
}

func (s *Oauth2ServiceV1) IssueAdminToken(ctx context.Context,
	req *pb.IssueAdminTokenRequest) (*pb.IssueTokenResponse, error) {
	passwordByte, _, err := s.kvOp.Get(ctx, model.KeyAdminPassword)
	if err != nil {
		log.Errorf("error get admin password: %s", err)
		return nil, pb.Oauth2ErrInternalStore()
	}
	password := string(passwordByte)
	if password != req.Password {
		log.Errorf("error admin password not match(%s -- %s)", password, req.Password)
		return nil, pb.Oauth2ErrPasswordNotMatch()
	}
	token, _, err := s.genToken("admin")
	if err != nil {
		log.Errorf("error issue(admin) oauth2 token gen plugin token: %s", err)
		return nil, pb.Oauth2ErrUnknown()
	}
	log.Debugf("issue(admin) oauth2 token: %s", token)
	return &pb.IssueTokenResponse{
		TokenType:   "Bearer",
		AccessToken: token,
		ExpiresIn:   int32(expires.Seconds()),
	}, nil
}

func (s *Oauth2ServiceV1) VerifyToken(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	header := transport_http.HeaderFromContext(ctx)
	token, ok := header[model.AuthorizationHeader]
	if !ok {
		log.Error("error get token \"\"")
		return nil, pb.Oauth2ErrInvalidToken()
	}
	_, valid, err := s.secretProvider.Parse(strings.TrimPrefix(token[0], "Bearer "))
	if err != nil {
		log.Errorf("error parse token(%s): %s", token, err)
		return nil, pb.Oauth2ErrInvalidToken()
	}
	if !valid {
		return nil, pb.Oauth2ErrInvalidToken()
	}
	return &emptypb.Empty{}, nil
}

func (s *Oauth2ServiceV1) UpdateAdminPassword(ctx context.Context, req *pb.UpdateAdminPasswordRequest) (*emptypb.Empty, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.Oauth2ErrUnknown()
	}
	if u.Tenant != model.TKeelTenant || u.User != model.TKeelUser {
		return nil, pb.Oauth2ErrPermissionDenied()
	}
	if !checkPassword(req.NewPassword) {
		return nil, pb.Oauth2ErrPasswordNotCompliant()
	}
	if err = s.kvOp.Delete(ctx, model.KeyAdminPassword); err != nil {
		log.Errorf("error delete rudder admin password: %s", err)
		return nil, pb.Oauth2ErrInternalStore()
	}
	baseStr := base64.StdEncoding.EncodeToString([]byte(req.NewPassword))
	if err = s.kvOp.Create(ctx, model.KeyAdminPassword,
		[]byte(baseStr)); err != nil {
		log.Errorf("error create rudder admin password: %s", err)
		return nil, pb.Oauth2ErrInternalStore()
	}
	log.Debugf("new admin password: %s -- %s", req.NewPassword, baseStr)
	return &emptypb.Empty{}, nil
}

func (s *Oauth2ServiceV1) checkPluginSecret(ps, os string) error {
	return nil
	// if ps == os {.
	// 	return nil.
	// }.
	// return ErrSecretNotMatch.
}

func (s *Oauth2ServiceV1) genToken(sub string, tokenKV ...string) (token, jti string, err error) {
	m := make(map[string]interface{})
	if len(tokenKV) != 0 {
		if len(tokenKV)%2 != 0 {
			err = errors.New("invalid token KV")
			return
		}
		for i := 0; i < len(tokenKV); i += 2 {
			m[tokenKV[i]] = tokenKV[i+1]
		}
	}
	duration := expires
	token, _, err = s.secretProvider.Token(sub, "", duration, m)
	if err != nil {
		err = fmt.Errorf("error token: %w", err)
		return
	}
	jti, ok := m["jti"].(string)
	if !ok {
		err = errors.New("error check")
		return
	}
	return
}

func (s *Oauth2ServiceV1) checkPluginWhiteList(id string) bool {
	for _, v := range s.pluginIDWhiteList {
		if v == id {
			return true
		}
	}
	return false
}

func checkPassword(password string) bool {
	return len(password) >= 6
}
