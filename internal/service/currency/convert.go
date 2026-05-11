package currency

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"finance_tracker/internal/entities"
	"finance_tracker/internal/utils"
)

const (
	legacyRateScale int64 = 100
	RateScale       int64 = 100000000
)

// ErrRatesUnavailable is returned when no currency rate data exists in memory or the database.
var ErrRatesUnavailable = errors.New("currency rates unavailable")

// ErrCurrencyNotInRates is returned when the requested currency has no rate entry in the available data.
var ErrCurrencyNotInRates = errors.New("currency not found in rates")

// ErrInvalidRate is returned when a persisted rate cannot be used for conversion.
var ErrInvalidRate = errors.New("invalid currency rate")

// ErrAmountTooLarge is returned when the converted USD amount cannot fit in int64.
var ErrAmountTooLarge = errors.New("amount too large")

// ConvertToUSD converts amountMinor (in the given currency) to USD minor units.
func (s *Service) ConvertToUSD(ctx context.Context, currency entities.Currency, amountMinor int64) (int64, error) {
	rates, err := s.ratesFor(ctx)
	if err != nil {
		return 0, err
	}

	rate, ok := rates.RateAgainstUSD[currency]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrCurrencyNotInRates, currency)
	}
	if rate <= 0 || rates.Scale <= 0 {
		return 0, fmt.Errorf("%w: %s", ErrInvalidRate, currency)
	}

	amountBig := big.NewInt(amountMinor)
	scaleBig := big.NewInt(rates.Scale)
	rateBig := big.NewInt(rate)

	product := new(big.Int).Mul(amountBig, scaleBig)
	converted := new(big.Int).Quo(product, rateBig)
	if !converted.IsInt64() {
		return 0, ErrAmountTooLarge
	}
	return converted.Int64(), nil
}

func (s *Service) ratesFor(ctx context.Context) (entities.CurrencyRate, error) {
	todayKey := utils.TimeToDayIntNum(time.Now())

	if rates, ok := s.currencies.Load(todayKey); ok {
		return normalizeRateScale(rates), nil
	}

	dbRow, err := s.repo.LoadLatestCurrencyRates(ctx)
	if err != nil {
		return entities.CurrencyRate{}, fmt.Errorf("load currency rates from db: %w", err)
	}
	if dbRow == nil {
		return entities.CurrencyRate{}, ErrRatesUnavailable
	}

	rates := entities.CurrencyRate{
		RateAgainstUSD: make(map[entities.Currency]int64, len(dbRow.CurrencyData)),
		Scale:          inferRateScale(dbRow.CurrencyData),
	}
	for currStr, val := range dbRow.CurrencyData {
		c, errC := entities.CurrencyFromString(currStr)
		if errC != nil {
			continue
		}
		rates.RateAgainstUSD[c] = val
	}

	s.currencies.Store(dbRow.TimestampNum, rates)
	return rates, nil
}

func inferRateScale(data map[string]int64) int64 {
	usdRate := data[entities.CurrencyUSD.String()]
	if usdRate <= legacyRateScale*10 {
		return legacyRateScale
	}
	return RateScale
}

func normalizeRateScale(rates entities.CurrencyRate) entities.CurrencyRate {
	if rates.Scale > 0 {
		return rates
	}
	rates.Scale = RateScale
	return rates
}
