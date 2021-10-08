package keel

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func ParsePluginID(payload string) (string, error) {
	b, err := jwt.DecodeSegment(payload)
	if err != nil {
		return "", err
	}
	pmap := make(map[string]interface{})
	err = json.Unmarshal(b, &pmap)
	if err != nil {
		return "", err
	}
	pID, ok := pmap["client_id"]
	if !ok {
		return "", nil
	}
	return pID.(string), nil
}

func GetPluginIDFromRequest(req *http.Request) (string, error) {
	pluginJwt := req.Header.Get("x-plugin-jwt")
	typeAndToken := strings.Split(pluginJwt, " ")
	if len(typeAndToken) != 2 {
		return "", nil
	}
	jwtList := strings.Split(typeAndToken[1], ".")
	if len(jwtList) != 3 {
		return "", nil
	}
	pid, err := ParsePluginID(jwtList[1])
	if err != nil {
		return "", err
	}
	return pid, nil
}
