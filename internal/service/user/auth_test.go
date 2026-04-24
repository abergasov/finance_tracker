package user_test

import (
	"finance_tracker/internal/entities"
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/utils"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuildUICallbackURLIncludesSuccessfulSessionHandoff(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)

	uID := uuid.NewString()
	token, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind:   utils.AuthTokenKind,
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: uID,
		Email:  "person@example.com",
		Locale: "en",
		Name:   "Person Example",
	})
	require.NoError(t, err)

	redirectURL, err := url.Parse(container.ServiceUser.BuildUICallbackURL(&entities.AuthSession{
		Token: token,
		User: entities.AuthUser{
			ID:    uID,
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
	require.Equal(t, uID, fragment.Get("id"))
	require.Equal(t, "person@example.com", fragment.Get("email"))
	require.Equal(t, "Person Example", fragment.Get("name"))

	homeData, err := container.ServiceUser.ServeHomePage(container.Ctx, fragment.Get("token"))
	require.NoError(t, err)
	require.Equal(t, uID, homeData.User.ID)
	require.Equal(t, "person@example.com", homeData.User.Email)
	require.Equal(t, "Person Example", homeData.User.Name)
}
