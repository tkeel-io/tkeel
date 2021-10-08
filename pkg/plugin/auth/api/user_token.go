package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/tkeel-io/tkeel/pkg/token"
)

const userTokenSecret = "FZyPg6lWTtCt0yCf0ZnDNWhHSt1rtOho"

var (
	idProvider token.IdProvider
)

func init() {
	idProvider = token.InitIdProvider([]byte(userTokenSecret), "", "")
}

func genUserToken(userID, tenantID, tokenID string) (token, jti string, err error) {
	m := make(map[string]interface{})
	m["uid"] = userID
	m["tid"] = tenantID
	duration := 12 * time.Hour
	token, err = idProvider.Token("user", tokenID, duration, &m)
	jti = m["jti"].(string)
	return
}

func parseUserToken(token string) (userID, tenantID string, err error) {
	var m = make(map[string]interface{})
	if token == "" {
		err = errors.New("invalid token")
		return
	}
	m, err = idProvider.Validate(token)
	if err != nil {
		return
	}
	userID = m["uid"].(string)
	tenantID = m["tid"].(string)
	return
}

func checkAuth(req *http.Request) error {
	authToken := req.Header.Get("authorization")
	uid, tid, err := parseUserToken(authToken)
	if err != nil {
		return err
	}
	req.Header.Set("uid", uid)
	req.Header.Set("tid", tid)
	return nil
}
