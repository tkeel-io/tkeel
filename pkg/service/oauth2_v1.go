package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/oauth2/v1"
	"github.com/tkeel-io/tkeel/pkg/model/passwd"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/token"
	"github.com/tkeel-io/tkeel/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrSecretNotMatch = errors.New("secret not match")

type Oauth2ServiceV1 struct {
	pb.UnimplementedOauth2Server

	passwdOp       passwd.Operator
	whiteList      []string
	pluginOp       plugin.Operator
	secretProvider token.Provider
}

func NewOauth2ServiceV1(passwdOp passwd.Operator, pOp plugin.Operator) *Oauth2ServiceV1 {
	secret := util.RandStringBytesMaskImpr(0, 10)
	return &Oauth2ServiceV1{
		passwdOp:       passwdOp,
		whiteList:      []string{"rudder", "keel", "core"},
		pluginOp:       pOp,
		secretProvider: token.InitProvider(secret, "", ""),
	}
}

func (s *Oauth2ServiceV1) AddPluginWhiteList(ctx context.Context,
	req *pb.AddPluginWhiteListRequest) (*emptypb.Empty, error) {
	if s.checkPluginWhiteList(req.PluginId) {
		return nil, fmt.Errorf("error duplicate add")
	}
	s.whiteList = append(s.whiteList, req.PluginId)
	return &emptypb.Empty{}, nil
}

func (s *Oauth2ServiceV1) IssuePluginToken(ctx context.Context, req *pb.IssuePluginTokenRequest) (*pb.IssueTokenResponse, error) {
	pluginID := req.ClientId
	if pluginID == "" {
		log.Errorf("error invaild plugin id: empty string")
		return nil, pb.Oauth2ErrInvaildPluginId()
	}
	if !s.checkPluginWhiteList(pluginID) {
		plugin, err := s.pluginOp.Get(ctx, pluginID)
		if err != nil {
			log.Errorf("error issue(%s) oauth2 token: %w", pluginID, err)
			return nil, pb.Oauth2ErrInternalStore()
		}
		if err = s.checkPluginSecret(plugin.Secret.Data, req.ClientSecret); err != nil {
			log.Errorf("error issue(%s) oauth2 token: %w", pluginID, err)
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
		AccessToken: token,
		ExpiresIn:   int32((24 * time.Hour).Seconds()),
	}, nil
}

func (s *Oauth2ServiceV1) IssueAdminToken(ctx context.Context,
	req *pb.IssueAdminTokenRequest) (*pb.IssueTokenResponse, error) {
	password, err := s.passwdOp.Get(ctx)
	if err != nil {
		log.Errorf("error get admin password: %s", err)
		return nil, pb.Oauth2ErrInternalStore()
	}
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
		AccessToken: token,
		ExpiresIn:   int32((24 * time.Hour).Seconds()),
	}, nil
}

func (s *Oauth2ServiceV1) checkPluginSecret(ps, os string) error {
	if ps == os {
		return nil
	}
	return ErrSecretNotMatch
}

func (s *Oauth2ServiceV1) genToken(sub string, tokenKV ...string) (token, jti string, err error) {
	m := make(map[string]interface{})
	if len(tokenKV) != 0 {
		if len(tokenKV)%2 != 0 {
			err = errors.New("invaild token KV")
			return
		}
		for i := 0; i < len(tokenKV); i += 2 {
			m[tokenKV[i]] = m[tokenKV[i+1]]
		}
	}
	duration := 24 * time.Hour
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
	for _, v := range s.whiteList {
		if v == id {
			return true
		}
	}
	return false
}
