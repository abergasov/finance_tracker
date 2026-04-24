package user

import (
	"context"
	"finance_tracker/internal/config"
	"finance_tracker/internal/logger"
	"finance_tracker/internal/repository"
	"finance_tracker/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Service struct {
	ctx   context.Context
	log   logger.AppLogger
	repo  *repository.Repo
	cfg   *config.AppConfig
	oauth *oauth2.Config
}

func InitService(ctx context.Context, log logger.AppLogger, repo *repository.Repo, cfg *config.AppConfig) *Service {
	utils.SetSignedKey(cfg.Auth.Token.SigningKey)
	return &Service{
		ctx:   ctx,
		repo:  repo,
		log:   log.With(logger.WithService("sampler")),
		cfg:   cfg,
		oauth: googleOAuthConfig(cfg),
	}
}

func (s *Service) Stop() {
	s.log.Info("stopping service")
}

func googleOAuthConfig(cfg *config.AppConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.Auth.Google.ClientID,
		ClientSecret: cfg.Auth.Google.ClientSecret,
		RedirectURL:  cfg.Auth.Google.RedirectURL,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
