package token

import (
	"time"
)

type IdProvider interface {
	Token(sub, jti string, d time.Duration, m *map[string]interface{}) (string, error)
	Validate(tokenStr string) (map[string]interface{}, error)
}
