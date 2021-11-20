/*
Copyright 2021 The tKeel Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

func InitProvider(secret []byte, rsaPriPath, rsaPubPath string) Provider {
	return NewBasicJWTIdentityProvider(secret, loadRSAPrivateKeyFromDisk(rsaPriPath), loadRSAPublicKeyFromDisk(rsaPubPath))
}

func loadRSAPrivateKeyFromDisk(location string) *rsa.PrivateKey {
	if location == "" {
		return nil
	}
	keyData, e := ioutil.ReadFile(location)
	if e != nil {
		panic(e.Error())
	}
	key, e := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if e != nil {
		panic(e.Error())
	}
	return key
}

func loadRSAPublicKeyFromDisk(location string) *rsa.PublicKey {
	if location == "" {
		return nil
	}
	keyData, e := ioutil.ReadFile(location)
	if e != nil {
		panic(e.Error())
	}
	key, e := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if e != nil {
		panic(e.Error())
	}
	return key
}

type BasicJWTIdentityProvider struct {
	secret []byte
	rsapri *rsa.PrivateKey
	rsapub *rsa.PublicKey
}

func NewBasicJWTIdentityProvider(secret []byte, priKey *rsa.PrivateKey, pubKey *rsa.PublicKey) *BasicJWTIdentityProvider {
	return &BasicJWTIdentityProvider{secret, priKey, pubKey}
}

func (idp *BasicJWTIdentityProvider) Token(sub, jti string, d time.Duration, m map[string]interface{}) (token string, expires int64, err error) {
	now := time.Now().UTC()
	exp := now.Add(d)
	if jti == "" {
		jti = uuid.New().String()
	}
	m["sub"] = sub
	m["iss"] = "manager"
	m["nbf"] = now
	m["aud"] = "keel"
	m["iat"] = now
	m["exp"] = exp
	m["jti"] = jti
	expires = exp.Unix()
	if idp.rsapri != nil {
		token, err = idp.rsaGen(m)
		return
	}

	token, err = idp.hsGen(m)
	return
}

func (idp *BasicJWTIdentityProvider) Validate(tokenStr string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if idp.rsapub != nil {
			return idp.rsapub, nil
		}
		return idp.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error jwt parsh with claims: %w", err)
	}
	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		return *claims, nil
	}
	return nil, ErrUnauthorizedAccess
}

func (idp *BasicJWTIdentityProvider) hsGen(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(idp.secret)
	if err != nil {
		return "", fmt.Errorf("error signed string: %w", err)
	}
	return t, nil
}

func (idp *BasicJWTIdentityProvider) rsaGen(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t, err := token.SignedString(idp.rsapri)
	if err != nil {
		return "", fmt.Errorf("error signed string: %w", err)
	}
	return t, nil
}
