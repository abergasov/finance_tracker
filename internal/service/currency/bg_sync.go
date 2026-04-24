package currency

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"finance_tracker/internal/entities"
	"finance_tracker/internal/utils"
)

func (s *Service) bgSyncCurrencies() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	s.SyncCurrencies()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.SyncCurrencies()
		}
	}
}

func (s *Service) SyncCurrencies() {
	now := utils.TimeToDayIntNum(time.Now())
	if _, ok := s.currencies.Load(now); ok {
		return
	}
	rates, err := s.getExchangeUSDRate(s.ctx, s.cfg.ExchangeToken, time.Now())
	if err != nil {
		s.log.Error("failed to get exchange rate", err)
		return
	}
	s.currencies.Store(now, *rates)
	if err = s.repo.SaveCurrencyRates(s.ctx, rates.RateAgainstUSD, time.Now()); err != nil {
		s.log.Error("failed to save currency rates", err)
	}
}

func (s *Service) getExchangeUSDRate(ctx context.Context, token string, date time.Time) (*entities.CurrencyRate, error) {
	type exchangeRateResponse struct {
		Disclaimer string             `json:"disclaimer"`
		License    string             `json:"license"`
		Timestamp  int                `json:"timestamp"`
		Base       string             `json:"base"`
		Rates      map[string]float64 `json:"rates"`
	}
	exchangeURL := fmt.Sprintf("https://openexchangerates.org/api/historical/%s.json?app_id=%s", date.Format(time.DateOnly), token)
	res, code, err := utils.GetCurl[exchangeRateResponse](ctx, exchangeURL, nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("failed to get exchange rate: %d", code)
	}
	result := &entities.CurrencyRate{
		RateAgainstUSD: make(map[entities.Currency]int64, 200),
	}
	for currencyStr, rate := range res.Rates {
		c, errC := entities.CurrencyFromString(currencyStr)
		if errC != nil {
			s.log.Error("unknown currency rate", errC)
			continue
		}

		// Store rate as fixed-point integer scaled by 100 (e.g., 1.235 -> 124).
		result.RateAgainstUSD[c] = int64(math.Round(rate * 100))
	}

	return result, nil
}
