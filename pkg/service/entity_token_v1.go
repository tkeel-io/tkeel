package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	pb "github.com/tkeel-io/tkeel/api/entity/v1"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
	"github.com/tkeel-io/kit/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EntityTokenService struct {
	EntityTokenOp TokenOperator
	pb.UnimplementedEntityTokenServer
}

func NewEntityTokenService(operator TokenOperator) *EntityTokenService {
	return &EntityTokenService{EntityTokenOp: operator}
}

func (s *EntityTokenService) CreateEntityToken(ctx context.Context, req *pb.CreateEntityTokenRequest) (*pb.CreateEntityTokenResponse, error) {
	now := time.Now()
	entity := &EToken{
		EntityID:   req.GetBody().GetEntityId(),
		EntityType: req.GetBody().GetEntityType(),
		Owner:      req.GetBody().GetOwner(),
		CreatedAt:  now.Unix(),
	}
	if req.GetBody().GetExpiresIn() == 0 {
		entity.ExpiredAt = now.Add(time.Hour * 24 * 365).Unix()
	} else {
		entity.ExpiredAt = now.Add(time.Second * time.Duration(req.GetBody().GetExpiresIn())).Unix()
	}

	token, err := s.EntityTokenOp.CreateToken(ctx, entity)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	return &pb.CreateEntityTokenResponse{Token: token}, nil
}

func (s *EntityTokenService) TokenInfo(ctx context.Context, req *pb.TokenInfoRequest) (*pb.TokenInfoResponse, error) {
	entity, err := s.EntityTokenOp.GetEntityInfo(ctx, req.GetToken())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}

	return &pb.TokenInfoResponse{
		EntityType: entity.EntityType,
		EntityId:   entity.EntityID,
		Owner:      entity.Owner,
		ExpiredAt:  entity.ExpiredAt,
		CreatedAt:  entity.CreatedAt,
	}, nil
}

func (s *EntityTokenService) DeleteEntityToken(ctx context.Context, req *pb.TokenInfoRequest) (*emptypb.Empty, error) {
	err := s.EntityTokenOp.DeleteToken(ctx, req.GetToken())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &emptypb.Empty{}, nil
}

type TokenOp struct {
	storeName string
	operator  dapr.Client
}

func NewEntityTokenOperator(storeName string, client dapr.Client) *TokenOp {
	if client == nil || storeName == "" {
		return nil
	}
	return &TokenOp{storeName: storeName, operator: client}
}

func (e *TokenOp) DeleteToken(ctx context.Context, key string) (err error) {
	err = e.operator.DeleteState(ctx, e.storeName, key)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("delete state  %w", err)
	}
	return
}

func (e *TokenOp) CreateToken(ctx context.Context, entity *EToken) (token string, err error) {
	value, err := json.Marshal(entity)
	if err != nil {
		return "", fmt.Errorf("marshal entity token %w", err)
	}
	i := 0
	key := entity.MD5ID(&i)
	resultKey := ""
	var item *dapr.StateItem
	for item, _ = e.operator.GetState(ctx, e.storeName, key); item.Value == nil && i < 4; key = entity.MD5ID(&i) {
		err = e.operator.SaveBulkState(ctx, e.storeName,
			&dapr.SetStateItem{
				Key:   key,
				Value: value,
				Etag:  &dapr.ETag{Value: item.Etag},
				Options: &dapr.StateOptions{
					Concurrency: dapr.StateConcurrencyFirstWrite,
					Consistency: dapr.StateConsistencyStrong,
				},
			})
		if err != nil {
			return "", fmt.Errorf("create token save state %w", err)
		}
		resultKey = key
		i = 4
	}
	if item.Value != nil {
		return "", errors.New("entity hash three repetitions on key ")
	}

	return resultKey, nil
}

func (e *TokenOp) GetEntityInfo(ctx context.Context, key string) (entity *EToken, err error) {
	item, err := e.operator.GetState(ctx, e.storeName, key)
	if err != nil {
		return nil, fmt.Errorf("entity info get state %w", err)
	}
	entity = &EToken{}
	if err = json.Unmarshal(item.Value, entity); err != nil {
		return nil, fmt.Errorf("unmarshal entity info  %w", err)
	}
	return
}

type EToken struct {
	EntityID   string `json:"entity_id"`
	EntityType string `json:"entity_type"`
	Owner      string `json:"owner"`
	CreatedAt  int64  `json:"created_at"`
	ExpiredAt  int64  `json:"expired_at"`
}

func (token *EToken) MD5ID(i *int) string {
	*i++
	buf := bytes.NewBufferString(token.EntityID)
	buf.WriteString(token.EntityType)
	buf.WriteString(token.Owner)
	buf.WriteString(strconv.FormatInt(token.CreatedAt, 10))
	access := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	access = strings.TrimRight(access, "=")
	return access
}

type TokenOperator interface {
	CreateToken(ctx context.Context, entity *EToken) (token string, err error)
	GetEntityInfo(ctx context.Context, token string) (entity *EToken, err error)
	DeleteToken(ctx context.Context, token string) (err error)
}
