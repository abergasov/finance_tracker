package currency

import (
	"context"
	"finance_tracker/internal/config"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/logger"
	"finance_tracker/internal/repository"
	"finance_tracker/internal/utils"
)

type Service struct {
	ctx context.Context
	log logger.AppLogger
	cfg *config.AppConfig

	repo       *repository.Repo
	currencies *utils.RWMap[int64, entities.CurrencyRate]
}

func NewService(ctx context.Context, log logger.AppLogger, cfg *config.AppConfig, repo *repository.Repo) *Service {
	return &Service{
		ctx:        ctx,
		log:        log,
		cfg:        cfg,
		repo:       repo,
		currencies: utils.NewRWMap[int64, entities.CurrencyRate](),
	}
}

func (s *Service) Run() {
	go s.bgSyncCurrencies()
}
