package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"time"
)

func GetEnv(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func RandomStr() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

func GetURLValue(u *url.URL, name string) string {
	return u.Query().Get(name)
}

func ReadBody2Json(b io.ReadCloser, a interface{}) error {
	defer b.Close()
	body, err := ioutil.ReadAll(b)
	if err != nil {
		err := fmt.Errorf("error read body: %s", err)
		return err
	}
	err = json.Unmarshal(body, a)
	if err != nil {
		err := fmt.Errorf("error json Unmarshal(%s): %s", string(body), err)
		return err
	}
	return nil
}
