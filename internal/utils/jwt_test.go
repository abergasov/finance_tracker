package utils_test

import (
	"finance_tracker/internal/entities"
	"finance_tracker/internal/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTokenRoundTrip(t *testing.T) {
	utils.SetSignedKey(uuid.NewString())
	token, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind:   utils.AuthTokenKind,
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: "user-1",
		Email:  "person@example.com",
		Name:   "Person Example",
		Locale: "en",
	})
	require.NoError(t, err)
	t.Logf("token: %s", token)

	user, err := utils.ParseAuthToken(token)
	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
	require.Equal(t, "person@example.com", user.Email)
	require.Equal(t, "Person Example", user.Name)
}

func TestVerifyOAuthStateTokenRejectsWrongKind(t *testing.T) {
	token, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind: utils.AuthTokenKind,
		Exp:  time.Now().Add(time.Hour).Unix(),
	})
	require.NoError(t, err)

	claim, err := utils.ParseClaims(token)
	require.NoError(t, err)
	require.Equal(t, utils.AuthTokenKind, claim.Kind)
}

func TestParseAuthTokenRejectsExpiredToken(t *testing.T) {
	token, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind:   utils.AuthTokenKind,
		Exp:    time.Now().Add(-time.Minute).Unix(),
		UserID: "user-1",
		Email:  "person@example.com",
	})
	require.NoError(t, err)

	_, err = utils.ParseAuthToken(token)
	require.Error(t, err)
}
