package util

import (
	"math/rand"
	"reflect"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	defaultSeed   = 123454321
	letterIdxBits = 6                    // 6 bits to represent a letter index.
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits.
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits.
)

func RandStringBytesMaskImpr(seed, n int64) []byte {
	if seed == 0 {
		seed = defaultSeed
	}
	r := rand.New(rand.NewSource(seed))
	if n <= 0 {
		n = r.Int63n(10) + 1
	}
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, r.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return b
}

func IsNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}
