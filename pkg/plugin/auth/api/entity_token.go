package api

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tkeel-io/tkeel/pkg/token"
)

var (
	entityIdp token.IDProvider
	devOnce   sync.Once
)

func InitEntityIdp(rsaPri, rsaPub string) {
	devOnce.Do(func() { entityIdp = token.InitIDProvider(nil, rsaPri, rsaPub) })
}

func genEntityToken(userID, tenantID, tokenID, entityID, entityType string, m map[string]interface{}) (token, jti string, err error) {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["uid"] = userID
	m["tid"] = tenantID
	m["eid"] = entityID
	m["typ"] = entityType
	duration := 365 * 24 * time.Hour
	token, err = entityIdp.Token("entity", tokenID, duration, m)
	if err != nil {
		err = fmt.Errorf("error token: %w", err)
		return
	}
	jti, ok := m["jti"].(string)
	if !ok {
		err = errors.New("error type assertion")
		return
	}
	return
}

func checkEntityToken(token string) error {
	if token == "" {
		return errors.New("invalid token")
	}
	_, err := entityIdp.Validate(token)
	if err != nil {
		return fmt.Errorf("error validate: %w", err)
	}
	return nil
}

func parseEntityToken(token string) (userID, tenantID, tokenID, entityID, entityType string, err error) {
	var m map[string]interface{}
	if token == "" {
		err = errors.New("invalid token")
		return
	}
	m, err = entityIdp.Validate(token)
	if err != nil {
		return
	}
	var ok bool
	userID, ok = m["uid"].(string)
	if !ok {
		err = errors.New("error type assertion")
		return
	}
	tenantID, ok = m["tid"].(string)
	if !ok {
		err = errors.New("error type assertion")
		return
	}
	tokenID, ok = m["jti"].(string)
	if !ok {
		err = errors.New("error type assertion")
		return
	}
	entityID, ok = m["eid"].(string)
	if !ok {
		err = errors.New("error type assertion")
		return
	}
	entityType, ok = m["typ"].(string)
	if !ok {
		err = errors.New("error type assertion")
		return
	}
	return userID, tenantID, tokenID, entityID, entityType, nil
}
