package token

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var (
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")
)

func InitIdProvider(secret []byte, rsaPriPath, rsaPubPath string) IdProvider {
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

func (idp *BasicJWTIdentityProvider) Token(sub, jti string, d time.Duration, m *map[string]interface{}) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(d)
	if jti == "" {
		jti = uuid.New().String()
	}
	(*m)["sub"] = sub
	(*m)["iss"] = "manager"
	(*m)["nbf"] = now
	(*m)["aud"] = "keel"
	(*m)["iat"] = now
	(*m)["exp"] = exp
	(*m)["jti"] = jti

	if idp.rsapri != nil {
		return idp.rsaGen(*m)
	}

	return idp.hsGen(*m)
}

func (idp *BasicJWTIdentityProvider) Validate(tokenStr string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if idp.rsapub != nil {
			return idp.rsapub, nil
		}
		return idp.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		return *claims, nil
	}
	return nil, ErrUnauthorizedAccess
}

func (idp *BasicJWTIdentityProvider) hsGen(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(idp.secret)
}

func (idp *BasicJWTIdentityProvider) rsaGen(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(idp.rsapri)
}
