package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/url"
	"os"
)

func GetEnv(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func RandomStr() string {
	bytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(26))
		if err != nil {
			n = big.NewInt(10)
		}
		b := n.Uint64() + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

func GetURLValue(u *url.URL, name string) string {
	return u.Query().Get(name)
}

func ReadBody2Json(b io.ReadCloser, a interface{}) error {
	body, err := ioutil.ReadAll(b)
	if err != nil {
		return fmt.Errorf("error read body: %w", err)
	}
	err = json.Unmarshal(body, a)
	if err != nil {
		return fmt.Errorf("error json Unmarshal(%s): %w", string(body), err)
	}
	return nil
}
