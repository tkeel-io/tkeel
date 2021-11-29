package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/oauth2/v1"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/token"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrSecretNotMatch = errors.New("secret not match")

type Oauth2ServiceV1 struct {
	pb.UnimplementedOauth2Server

	secret         string
	whiteList      []string
	pluginOp       plugin.Operator
	secretProvider token.Provider
}

func NewOauth2ServiceV1(secret string, pluginOperator plugin.Operator) *Oauth2ServiceV1 {
	return &Oauth2ServiceV1{
		secret:         secret,
		whiteList:      []string{"rudder", "keel", "core"},
		pluginOp:       pluginOperator,
		secretProvider: token.InitProvider([]byte(secret), "", ""),
	}
}

func (s *Oauth2ServiceV1) AddWhiteList(ctx context.Context,
	req *pb.AddWhiteListRequest) (*emptypb.Empty, error) {
	if req.Secret != s.secret {
		return nil, fmt.Errorf("error secret(%s) not match", req.Secret)
	}
	if s.checkWhiteList(req.ClientId) {
		return nil, fmt.Errorf("error duplicate add")
	}
	s.whiteList = append(s.whiteList, req.ClientId)
	return &emptypb.Empty{}, nil
}

func (s *Oauth2ServiceV1) IssueOauth2Token(ctx context.Context, req *pb.IssueOauth2TokenRequest) (*pb.IssueOauth2TokenResponse, error) {
	pluginID := req.ClientId
	if pluginID == "" {
		log.Errorf("error invaild plugin id: empty string")
		return nil, pb.ErrInvaildPluginId()
	}
	if !s.checkWhiteList(pluginID) {
		plugin, err := s.pluginOp.Get(ctx, pluginID)
		if err != nil {
			log.Errorf("error issue(%s) oauth2 token: %w", pluginID, err)
			return nil, pb.ErrInternalStore()
		}
		if err = s.checkPluginSecret(plugin.Secret, req.ClientSecret); err != nil {
			log.Errorf("error issue(%s) oauth2 token: %w", pluginID, err)
			return nil, pb.ErrSecretNotMatch()
		}
	}
	token, _, err := s.genPluginToken(pluginID)
	if err != nil {
		log.Errorf("error issue(%s) oauth2 token gen plugin token: %s", pluginID, err)
		return nil, pb.ErrUnknown()
	}

	return &pb.IssueOauth2TokenResponse{
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

func (s *Oauth2ServiceV1) genPluginToken(pID string) (token, jti string, err error) {
	m := make(map[string]interface{})
	m["plugin_id"] = pID
	duration := 24 * time.Hour
	token, _, err = s.secretProvider.Token("user", "", duration, m)
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

func (s *Oauth2ServiceV1) checkWhiteList(id string) bool {
	for _, v := range s.whiteList {
		if v == id {
			return true
		}
	}
	return false
}
