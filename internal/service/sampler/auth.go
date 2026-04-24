package sampler

import (
	"context"
	"errors"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/logger"
	"finance_tracker/internal/utils"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type googleUserInfo struct {
	Email         string `json:"email"`
	Locale        string `json:"locale"`
	Name          string `json:"name"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (s *Service) StartGoogleAuth() (redirectURL, state string, err error) {
	claims, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind:  utils.OauthStateKind,
		Exp:   time.Now().Add(10 * time.Minute).Unix(),
		Nonce: utils.RandomBase64(16),
	})
	if err != nil {
		return "", "", fmt.Errorf("sign oauth state: %w", err)
	}

	return s.oauth.AuthCodeURL(claims, oauth2.AccessTypeOffline), claims, nil
}

func (s *Service) VerifyOAuthStateToken(state string) error {
	claims, err := utils.ParseClaims(state)
	if err != nil {
		return err
	}
	if claims.Kind != utils.OauthStateKind {
		return errors.New("invalid oauth state")
	}
	return nil
}

func (s *Service) CompleteGoogleAuth(ctx context.Context, code string) (*entities.AuthSession, error) {
	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange google code: %w", err)
	}

	googleUser, err := s.fetchGoogleUserInfo(ctx, token)
	if err != nil {
		return nil, err
	}
	if googleUser.Email == "" || !googleUser.VerifiedEmail {
		return nil, fmt.Errorf("google account email is missing or not verified")
	}

	user, err := s.repo.SaveUserByEmail(ctx, googleUser.Email, googleUser.Name)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	sessionToken, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind:   utils.AuthTokenKind,
		Exp:    time.Now().Add(s.cfg.Auth.Token.ExpirationMinutes * time.Minute).Unix(),
		UserID: user.ID.String(),
		Email:  user.Email,
		Name:   user.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("sign auth token: %w", err)
	}

	return &entities.AuthSession{
		Token: sessionToken,
		User: entities.AuthUser{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}

func (s *Service) ParseAuthToken(token string) (*entities.AuthUser, error) {
	return utils.ParseAuthToken(token)
}

func (s *Service) BuildUICallbackURL(session *entities.AuthSession, authErr error) string {
	baseURL := strings.TrimRight(s.cfg.Auth.UIBaseURL, "/")
	if baseURL == "" {
		baseURL = "/"
	}

	callbackURL := baseURL
	if !strings.HasSuffix(callbackURL, "/auth/callback") {
		callbackURL += "/auth/callback"
	}

	params := url.Values{}
	if authErr != nil {
		params.Set("error", authErr.Error())
	} else if session != nil {
		params.Set("token", session.Token)
		params.Set("id", session.User.ID)
		params.Set("email", session.User.Email)
		params.Set("name", session.User.Name)
	}

	if encoded := params.Encode(); encoded != "" {
		return callbackURL + "#" + encoded
	}
	return callbackURL
}

func (s *Service) UICORSOrigin() string {
	parsedURL, err := url.Parse(strings.TrimSpace(s.cfg.Auth.UIBaseURL))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return ""
	}
	return parsedURL.Scheme + "://" + parsedURL.Host
}

func (s *Service) IsSecureCookieEnabled() bool {
	return s.cfg.Auth.SecureCookies
}

func (s *Service) fetchGoogleUserInfo(parentCtx context.Context, token *oauth2.Token) (*googleUserInfo, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()
	userInfo, code, err := utils.GetCurl[googleUserInfo](ctx, "https://www.googleapis.com/oauth2/v2/userinfo", map[string]string{
		"Authorization": "Bearer " + token.AccessToken,
	})
	if code != http.StatusOK {
		s.log.Error("failed to get user info from google", err, logger.WithStatusCode(code))
		return nil, fmt.Errorf("failed to get google user info")
	}
	if err != nil {
		s.log.Error("failed to fetch user info from google", err, logger.WithStatusCode(code))
		return nil, fmt.Errorf("failed to fetch user info from google")
	}
	return userInfo, nil
}
