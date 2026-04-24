package testhelpers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/logger"
	"finance_tracker/internal/routes"
	"finance_tracker/internal/test_helpers/seed"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestServer struct {
	t         *testing.T
	appPort   int
	client    http.Client
	authToken string
	container *TestContainer
}

func NewTestServerWithUser(t *testing.T, container *TestContainer) *TestServer {
	user := seed.NewUserBuilder().PopulateTest(t, container.Repo)
	srv := NewTestServer(t, container)
	srv.AuthUser(user.Email)
	return srv
}

func NewTestServer(t *testing.T, container *TestContainer) *TestServer {
	appPort := GetFreePort(t)
	srv := &TestServer{
		t:         t,
		appPort:   appPort,
		client:    *http.DefaultClient,
		container: container,
	}

	appLog := logger.NewAppSLogger()
	appHTTPServer := routes.InitAppRouter(appLog, container.ServiceUser, fmt.Sprintf(":%d", srv.appPort), false)
	t.Cleanup(func() {
		require.NoError(t, appHTTPServer.Stop())
	})
	go appHTTPServer.Run() //nolint:errcheck // ignore error as we are not interested in it
	srv.waitForReady(t)
	return srv
}

func (ts *TestServer) AuthUser(mail string) {
	usr, err := ts.container.Repo.GetUserByEmail(ts.container.Ctx, mail)
	require.NoError(ts.t, err, "get user by email")
	require.NotNil(ts.t, usr, "get user by email")
	ts.authToken = signToken(ts.t, ts.container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
}

func (ts *TestServer) ResetUser() {
	ts.authToken = ""
}

func (ts *TestServer) DisableRedirects() {
	ts.client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (ts *TestServer) Get(t *testing.T, path string) *TestResponse {
	t.Helper()
	return ts.Request(t, http.MethodGet, path, nil, nil, nil)
}

func (ts *TestServer) GetWithCookie(t *testing.T, path string, cookies map[string]string) *TestResponse {
	t.Helper()
	return ts.Request(t, http.MethodGet, path, nil, nil, cookies)
}

func (ts *TestServer) GetWithHeader(t *testing.T, path string, headers map[string]string) *TestResponse {
	t.Helper()
	return ts.Request(t, http.MethodGet, path, nil, headers, nil)
}

func (ts *TestServer) Post(t *testing.T, path string, body any) *TestResponse {
	t.Helper()
	return ts.Request(t, http.MethodPost, path, body, nil, nil)
}

func (ts *TestServer) Put(t *testing.T, path string, body any) *TestResponse {
	t.Helper()
	return ts.Request(t, http.MethodPut, path, body, nil, nil)
}

func (ts *TestServer) Delete(t *testing.T, path string, body any) *TestResponse {
	t.Helper()
	return ts.Request(t, http.MethodDelete, path, body, nil, nil)
}

func (ts *TestServer) Request(t *testing.T, method, path string, body interface{}, headers, cookies map[string]string) *TestResponse {
	t.Helper()

	var b []byte
	var err error
	if body != nil {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Content-Type"] = "application/json"
		b, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}

	u := fmt.Sprintf("http://localhost:%d%s", ts.appPort, path)
	req, err := http.NewRequest(method, u, bytes.NewBuffer(b))
	require.NoError(t, err, "failed to construct new request for url %s: %s", u, err)
	if err != nil {
		t.Fatal(err)
	}

	if len(headers) > 0 {
		for headerKey, headerVal := range headers {
			req.Header.Add(headerKey, headerVal)
		}
	}
	if len(cookies) > 0 {
		for cookieKey, cookieVal := range cookies {
			req.AddCookie(&http.Cookie{
				Name:  cookieKey,
				Value: cookieVal,
			})
		}
	}
	if ts.authToken != "" {
		req.Header.Add("Authorization", fmt.Sprint("Bearer ", ts.authToken))
	}

	res, err := ts.client.Do(req)
	require.NoError(t, err, "failed to make request to %s: %s", u, err)
	t.Cleanup(func() {
		require.NoError(t, res.Body.Close())
	})
	return &TestResponse{Res: res}
}

func (ts *TestServer) waitForReady(t testing.TB) {
	t.Helper()
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", ts.appPort), 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		return false
	}, 5*time.Second, 10*time.Millisecond, "failed to start test HTTP server")
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
