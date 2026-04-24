package routes

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const GoogleOAuthStateCookie = "google_oauth_state"

func (s *Router) handleGoogleLogin(ctx fiber.Ctx) error {
	url, state, err := s.service.StartGoogleAuth()
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "failed to start google auth"})
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     GoogleOAuthStateCookie,
		Value:    state,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   s.service.IsSecureCookieEnabled(),
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
	})
	return ctx.Redirect().To(url)
}

func (s *Router) handleGoogleCallback(ctx fiber.Ctx) error {
	state := ctx.Query("state")
	cookieState := ctx.Cookies(GoogleOAuthStateCookie)
	if state == "" || cookieState == "" || state != cookieState {
		return s.redirectOrRespondAuthFailure(ctx, errors.New("invalid state"), http.StatusBadRequest)
	}
	if err := s.service.VerifyOAuthStateToken(state); err != nil {
		return s.redirectOrRespondAuthFailure(ctx, err, http.StatusBadRequest)
	}

	session, err := s.service.CompleteGoogleAuth(ctx.Context(), ctx.Query("code"))
	if err != nil {
		return s.redirectOrRespondAuthFailure(ctx, err, http.StatusUnauthorized)
	}

	s.clearOAuthStateCookie(ctx)
	return ctx.Redirect().To(s.service.BuildUICallbackURL(session, nil))
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

func (s *Router) handleCurrentUser(ctx fiber.Ctx) error {
	token, ok := bearerToken(ctx.Get("Authorization"))
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
	}

	user, err := s.service.ServeHomePage(ctx.Context(), token)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}
	return ctx.JSON(user)
}

func (s *Router) redirectOrRespondAuthFailure(ctx fiber.Ctx, err error, statusCode int) error {
	s.clearOAuthStateCookie(ctx)
	return ctx.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
}

func (s *Router) clearOAuthStateCookie(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     GoogleOAuthStateCookie,
		Value:    "",
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   s.service.IsSecureCookieEnabled(),
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})
}
