package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"finance_tracker/internal/entities"
	"fmt"
	"strings"
	"time"
)

const (
	AuthTokenKind  = "auth"
	OauthStateKind = "oauth_state"
)

var (
	signedKey       string
	ErrInvalidToken = errors.New("invalid token")
)

func SetSignedKey(key string) {
	signedKey = key
}

func SignClaims(claims *entities.SignedTokenClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal token claims: %w", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	encodedSignature := base64.RawURLEncoding.EncodeToString(tokenSignature(encodedPayload))
	return encodedPayload + "." + encodedSignature, nil
}

func ParseAuthToken(token string) (*entities.AuthUser, error) {
	claims, err := ParseClaims(token)
	if err != nil {
		return nil, err
	}
	if claims.Kind != AuthTokenKind || claims.UserID == "" || claims.Email == "" {
		return nil, ErrInvalidToken
	}

	return &entities.AuthUser{
		ID:    claims.UserID,
		Email: claims.Email,
		Name:  claims.Name,
	}, nil
}

func ParseClaims(token string) (*entities.SignedTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !hmac.Equal(signature, tokenSignature(parts[0])) {
		return nil, ErrInvalidToken
	}

	claims := &entities.SignedTokenClaims{}
	if err = json.Unmarshal(payload, claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Exp <= time.Now().Unix() {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func tokenSignature(payload string) []byte {
	mac := hmac.New(sha256.New, []byte(signedKey))
	_, _ = mac.Write([]byte(payload))
	return mac.Sum(nil)
}
