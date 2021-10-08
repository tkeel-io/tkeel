package api

import (
	"errors"
	"sync"
	"time"

	"github.com/tkeel-io/tkeel/pkg/token"
)

var (
	entityIdp token.IdProvider
	devOnce   sync.Once
)

func InitEntityIdp(rsaPri, rsaPub string) {
	devOnce.Do(func() { entityIdp = token.InitIdProvider(nil, rsaPri, rsaPub) })
}

func genEntityToken(userID, tenantID, tokenID, entityID, entityType string, m *map[string]interface{}) (token, jti string, err error) {
	if m == nil {
		mm := make(map[string]interface{})
		m = &mm
	}
	(*m)["uid"] = userID
	(*m)["tid"] = tenantID
	(*m)["eid"] = entityID
	(*m)["typ"] = entityType
	duration := 365 * 24 * time.Hour
	token, err = entityIdp.Token("entity", tokenID, duration, m)
	jti = (*m)["jti"].(string)
	return
}

func parseEntityToken(token string) (userID, tenantID, tokenID, entityID, entityType string, err error) {
	var m = make(map[string]interface{})
	if token == "" {
		err = errors.New("invalid token")
		return
	}
	m, err = entityIdp.Validate(token)
	if err != nil {
		return
	}
	userID = m["uid"].(string)
	tenantID = m["tid"].(string)
	tokenID = m["jti"].(string)
	entityID = m["eid"].(string)
	entityType = m["typ"].(string)
	return
}
