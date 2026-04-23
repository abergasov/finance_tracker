package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

func RandomBase64(size int) string {
	randomBytes := make([]byte, size)
	if _, err := rand.Read(randomBytes); err != nil {
		return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes)
}
