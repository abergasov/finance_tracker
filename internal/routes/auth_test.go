package routes_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"finance_tracker/internal/entities"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"finance_tracker/internal/routes"
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"

	"github.com/stretchr/testify/require"
)

func TestGoogleLoginRedirectSetsStateCookie(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)
	srv.DisableRedirects()

	// when
	response := srv.Get(t, "/api/auth/google/login").RequireSeeOther(t)

	// then
	location, err := url.Parse(response.Res.Header.Get("Location"))
	require.NoError(t, err)
	require.Equal(t, "accounts.google.com", location.Host)
	require.Equal(t, "/o/oauth2/auth", location.Path)
	require.Equal(t, container.Cfg.Auth.Google.ClientID, location.Query().Get("client_id"))
	require.Equal(t, container.Cfg.Auth.Google.RedirectURL, location.Query().Get("redirect_uri"))

	state := location.Query().Get("state")
	require.NotEmpty(t, state)

	stateCookie := response.FindCookie(t, routes.GoogleOAuthStateCookie)
	require.Equal(t, state, stateCookie.Value)
	require.True(t, stateCookie.HttpOnly)
	require.Equal(t, "/", stateCookie.Path)
	require.WithinDuration(t, time.Now().Add(10*time.Minute), stateCookie.Expires, time.Minute)
}

func TestGoogleCallbackRejectsMismatchedState(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	response := srv.GetWithCookie(t, "/api/auth/google/callback?state=query-state", map[string]string{
		routes.GoogleOAuthStateCookie: "cookie-state",
	}).RequireBadRequest(t)
	require.Equal(t, "", response.Res.Header.Get("Location"))
	stateCookie := response.FindCookie(t, routes.GoogleOAuthStateCookie)
	require.Empty(t, stateCookie.Value)
	require.True(t, stateCookie.Expires.Before(time.Now().Add(time.Second)))
}

func TestCurrentUserAuthChecks(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	// Create a real user so ServeHomePage can load the DB row for default_currency.
	dbUser := seed.NewUserBuilder().PopulateTest(t, container.Repo)
	uID := dbUser.ID.String()

	validToken := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: uID,
		Email:  dbUser.Email,
		Locale: "en",
		Name:   dbUser.Name,
	})
	expiredToken := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(-time.Minute).Unix(),
		UserID: "user-1",
		Email:  "person@example.com",
	})

	t.Run("missing bearer token", func(t *testing.T) {
		srv.Get(t, "/api/v1/me").RequireUnauthorized(t)
	})

	t.Run("expired token", func(t *testing.T) {
		srv.GetWithHeader(t, "/api/v1/me", withBearer(expiredToken)).RequireUnauthorized(t)
	})

	t.Run("tampered token", func(t *testing.T) {
		srv.GetWithHeader(t, "/api/v1/me", withBearer(tamperToken(validToken))).RequireUnauthorized(t)
	})

	t.Run("valid token", func(t *testing.T) {
		var payload entities.HomePage
		srv.GetWithHeader(t, "/api/v1/me", withBearer(validToken)).RequireOk(t).RequireUnmarshal(t, &payload)

		require.Equal(t, dbUser.Email, payload.User.Email)
		require.Equal(t, dbUser.Name, payload.User.Name)
		require.Equal(t, uID, payload.User.ID)
		require.Equal(t, "USD", payload.User.DefaultCurrency)
	})
}

func TestCurrentUserCORSAllowsConfiguredUIOrigin(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	// Create a real user so ServeHomePage can load the DB row.
	dbUser := seed.NewUserBuilder().PopulateTest(t, container.Repo)
	validToken := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: dbUser.ID.String(),
		Email:  dbUser.Email,
	})

	t.Run("preflight allows configured ui origin", func(t *testing.T) {
		response := srv.Request(t, http.MethodOptions, "/api/v1/me", nil, map[string]string{
			"Origin":                         container.Cfg.Auth.UIBaseURL,
			"Access-Control-Request-Method":  http.MethodGet,
			"Access-Control-Request-Headers": "authorization",
		}, nil).RequireNoContent(t)
		require.Equal(t, container.Cfg.Auth.UIBaseURL, response.Res.Header.Get("Access-Control-Allow-Origin"))
		require.Contains(t, response.Res.Header.Get("Access-Control-Allow-Methods"), http.MethodGet)
		require.Contains(t, response.Res.Header.Get("Access-Control-Allow-Headers"), "Authorization")
		require.Contains(t, response.Res.Header.Get("Vary"), "Origin")
	})

	t.Run("authorized request includes allow origin header", func(t *testing.T) {
		response := srv.GetWithHeader(t, "/api/v1/me", map[string]string{
			"Origin":        container.Cfg.Auth.UIBaseURL,
			"Authorization": "Bearer " + validToken,
		}).RequireOk(t)
		require.Equal(t, container.Cfg.Auth.UIBaseURL, response.Res.Header.Get("Access-Control-Allow-Origin"))
		require.Contains(t, response.Res.Header.Get("Vary"), "Origin")
	})
}

func TestUpdateUserDefaultCurrency(t *testing.T) {
	container := testhelpers.GetClean(t)
	usr := seed.NewUserBuilder().PopulateTest(t, container.Repo)
	srv := testhelpers.NewTestServer(t, container)
	srv.AuthUser(usr.Email)

	t.Run("missing token returns 401", func(t *testing.T) {
		srv.ResetUser()
		defer srv.AuthUser(usr.Email)
		srv.Put(t, "/api/v1/me/currency", map[string]string{"currency": "EUR"}).RequireUnauthorized(t)
	})

	t.Run("invalid currency returns 400", func(t *testing.T) {
		srv.Put(t, "/api/v1/me/currency", map[string]string{"currency": "NOTREAL"}).RequireStatus(t, http.StatusBadRequest)
	})

	t.Run("valid currency is persisted and reflected in /me", func(t *testing.T) {
		srv.Put(t, "/api/v1/me/currency", map[string]string{"currency": "EUR"}).RequireOk(t)

		var homePage entities.HomePage
		srv.Get(t, "/api/v1/me").RequireOk(t).RequireUnmarshal(t, &homePage)
		require.Equal(t, "EUR", homePage.User.DefaultCurrency)
	})

	t.Run("token for non-existent user returns 404", func(t *testing.T) {
		// Sign a valid, non-expired token whose user ID does not exist in the DB.
		ghostToken := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
			Kind:   "auth",
			Exp:    time.Now().Add(time.Hour).Unix(),
			UserID: "00000000-0000-0000-0000-000000000000",
			Email:  "ghost@example.com",
		})
		srv.ResetUser()
		defer srv.AuthUser(usr.Email)
		srv.Request(t, http.MethodPut, "/api/v1/me/currency",
			map[string]string{"currency": "EUR"},
			withBearer(ghostToken),
			nil,
		).RequireOk(t)
	})

	t.Run("currency can be changed back", func(t *testing.T) {
		srv.Put(t, "/api/v1/me/currency", map[string]string{"currency": "USD"}).RequireOk(t)

		var homePage entities.HomePage
		srv.Get(t, "/api/v1/me").RequireOk(t).RequireUnmarshal(t, &homePage)
		require.Equal(t, "USD", homePage.User.DefaultCurrency)
	})
}

func signToken(t *testing.T, signingKey string, claims *entities.SignedTokenClaims) string {
	t.Helper()

	payload, err := json.Marshal(claims)
	require.NoError(t, err)

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(signingKey))
	_, err = mac.Write([]byte(encodedPayload))
	require.NoError(t, err)

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature
}

func tamperToken(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return token + "x"
	}

	parts[1] = parts[0]
	return strings.Join(parts, ".")
}

func withBearer(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}
