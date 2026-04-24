package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func RandomBase64(size int) string {
	randomBytes := make([]byte, size)
	if _, err := rand.Read(randomBytes); err != nil {
		panic("crypto/rand.Read failed: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes)
}
