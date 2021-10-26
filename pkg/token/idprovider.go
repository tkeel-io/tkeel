package token

import (
	"time"
)

type IDProvider interface {
	Token(sub, jti string, d time.Duration, m map[string]interface{}) (string, int64, error)
	Validate(tokenStr string) (map[string]interface{}, error)
}
