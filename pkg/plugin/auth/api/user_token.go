package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/tkeel-io/tkeel/pkg/token"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

var userTokenSecret = utils.GetEnv("USER_TOKEN_SECRET", "FZyPg6lWTtCt0yCf0ZnDNWhHSt1rtOho")

var (
	idProvider token.IDProvider
)

func init() {
	idProvider = token.InitIDProvider([]byte(userTokenSecret), "", "")
}

func genUserToken(userID, tenantID, tokenID string) (token, jti string, err error) {
	m := make(map[string]interface{})
	m["uid"] = userID
	m["tid"] = tenantID
	duration := 12 * time.Hour
	token, err = idProvider.Token("user", tokenID, duration, m)
	var ok bool
	jti, ok = m["jti"].(string)
	if !ok {
		err = errors.New("type assertion faild")
		return
	}
	return
}

func parseUserToken(token string) (userID, tenantID string, err error) {
	var m map[string]interface{}
	if token == "" {
		err = errors.New("invalid token")
		return
	}
	m, err = idProvider.Validate(token)
	if err != nil {
		return
	}
	var ok bool
	userID, ok = m["uid"].(string)
	if !ok {
		err = errors.New("type assertion faild")
		return
	}
	tenantID, ok = m["tid"].(string)
	if !ok {
		err = errors.New("type assertion faild")
		return
	}
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
