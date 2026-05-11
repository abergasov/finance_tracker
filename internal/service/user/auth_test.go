package user_test

import (
	"finance_tracker/internal/entities"
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"
	"finance_tracker/internal/utils"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildUICallbackURLIncludesSuccessfulSessionHandoff(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	dbUser := seed.NewUserBuilder().PopulateTest(t, container.Repo)

	token, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind:   utils.AuthTokenKind,
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: dbUser.ID.String(),
		Email:  dbUser.Email,
		Locale: "en",
		Name:   dbUser.Name,
	})
	require.NoError(t, err)

	redirectURL, err := url.Parse(container.ServiceUser.BuildUICallbackURL(&entities.AuthSession{
		Token: token,
		User: entities.AuthUser{
			ID:              dbUser.ID.String(),
			Email:           dbUser.Email,
			Name:            dbUser.Name,
			DefaultCurrency: dbUser.DefaultCurrency,
		},
	}, nil))
	require.NoError(t, err)
	require.Equal(t, "http://localhost:3000", redirectURL.Scheme+"://"+redirectURL.Host)
	require.Equal(t, "/auth/callback", redirectURL.Path)

	fragment, err := url.ParseQuery(redirectURL.Fragment)
	require.NoError(t, err)
	require.Equal(t, token, fragment.Get("token"))
	require.Equal(t, dbUser.ID.String(), fragment.Get("id"))
	require.Equal(t, dbUser.Email, fragment.Get("email"))
	require.Equal(t, dbUser.Name, fragment.Get("name"))
	require.Equal(t, dbUser.DefaultCurrency, fragment.Get("default_currency"))

	homeData, err := container.ServiceUser.ServeHomePage(container.Ctx, fragment.Get("token"))
	require.NoError(t, err)
	require.Equal(t, dbUser.ID.String(), homeData.User.ID)
	require.Equal(t, dbUser.Email, homeData.User.Email)
	require.Equal(t, dbUser.Name, homeData.User.Name)
	require.Equal(t, dbUser.DefaultCurrency, homeData.User.DefaultCurrency)
}
