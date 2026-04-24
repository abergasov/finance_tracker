package sampler_test

import (
	"finance_tracker/internal/entities"
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/utils"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildUICallbackURLIncludesSuccessfulSessionHandoff(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)

	token, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind:   utils.AuthTokenKind,
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: "user-1",
		Email:  "person@example.com",
		Locale: "en",
		Name:   "Person Example",
	})
	require.NoError(t, err)

	redirectURL, err := url.Parse(container.ServiceSampler.BuildUICallbackURL(&entities.AuthSession{
		Token: token,
		User: entities.AuthUser{
			ID:    "user-1",
			Email: "person@example.com",
			Name:  "Person Example",
		},
	}, nil))
	require.NoError(t, err)
	require.Equal(t, "http://localhost:3000", redirectURL.Scheme+"://"+redirectURL.Host)
	require.Equal(t, "/auth/callback", redirectURL.Path)

	fragment, err := url.ParseQuery(redirectURL.Fragment)
	require.NoError(t, err)
	require.Equal(t, token, fragment.Get("token"))
	require.Equal(t, "user-1", fragment.Get("id"))
	require.Equal(t, "person@example.com", fragment.Get("email"))
	require.Equal(t, "Person Example", fragment.Get("name"))

	user, err := container.ServiceSampler.ParseAuthToken(fragment.Get("token"))
	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
	require.Equal(t, "person@example.com", user.Email)
	require.Equal(t, "Person Example", user.Name)
}
