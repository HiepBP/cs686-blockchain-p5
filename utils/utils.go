package utils

import (
	"math/rand"
	"time"
)

func RandomHex(n int) ([]byte, error) {
	bytes := make([]byte, n/2)
	rand.Seed(time.Now().UnixNano())
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}
