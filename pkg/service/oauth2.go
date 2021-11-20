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
)

var ErrSecretNotMatch = errors.New("secret not match")

type Oauth2Service struct {
	pb.UnimplementedOauth2Server

	pluginOp       plugin.Operator
	secretProvider token.Provider
}

func NewOauth2Service(secret string, pluginOperator plugin.Operator) *Oauth2Service {
	return &Oauth2Service{
		pluginOp:       pluginOperator,
		secretProvider: token.InitProvider([]byte(secret), "", ""),
	}
}

func (s *Oauth2Service) IssueOauth2Token(ctx context.Context, req *pb.IssueOauth2TokenRequest) (*pb.IssueOauth2TokenResponse, error) {
	pluginID := req.ClientId
	if pluginID == "" {
		err := fmt.Errorf("error invaild plugin id: empty string")
		log.Error(err)
		return nil, err
	}
	plugin, err := s.pluginOp.Get(ctx, pluginID)
	if err != nil {
		err = fmt.Errorf("error issue(%s) oauth2 token: %w", pluginID, err)
		log.Error(err)
		return nil, err
	}
	if err = s.checkPluginSecret(plugin.Secret, req.ClientSecret); err != nil {
		err = fmt.Errorf("error issue(%s) oauth2 token: %w", pluginID, err)
		log.Error(err)
		return nil, err
	}
	token, _, err := s.genPluginToken(pluginID)
	if err != nil {
		err = fmt.Errorf("error issue(%s) oauth2 token gen plugin token: %w", pluginID, err)
		log.Error(err)
		return nil, err
	}

	return &pb.IssueOauth2TokenResponse{
		AccessToken: token,
		ExpiresIn:   int32((24 * time.Hour).Seconds()),
	}, nil
}

func (s *Oauth2Service) checkPluginSecret(ps, os string) error {
	if ps == os {
		return nil
	}
	return ErrSecretNotMatch
}

func (s *Oauth2Service) genPluginToken(pID string) (token, jti string, err error) {
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
