package currency_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"finance_tracker/internal/entities"
	"finance_tracker/internal/service/currency"
	testhelpers "finance_tracker/internal/test_helpers"

	"github.com/stretchr/testify/require"
)

func TestCurrencySync(t *testing.T) {
	// given
	testhelpers.SkipCIRun(t)
	container := testhelpers.GetClean(t)
	rates, err := container.Repo.LoadAllCurrencies(container.Ctx)
	require.NoError(t, err)
	require.Empty(t, rates)

	// when
	container.ServiceCurrency.SyncCurrencies()

	// then
	rates, err = container.Repo.LoadAllCurrencies(container.Ctx)
	require.NoError(t, err)
	require.Len(t, rates, 1)
}

func TestConvertToUSDUsesLatestStoredRates(t *testing.T) {
	container := testhelpers.GetClean(t)
	require.NoError(t, container.Repo.SaveCurrencyRates(container.Ctx, map[entities.Currency]int64{
		entities.CurrencyUSD: currency.RateScale,
		entities.CurrencyEUR: 108000000,
	}, time.Now()))

	amountMinorUSD, err := container.ServiceCurrency.ConvertToUSD(container.Ctx, entities.CurrencyEUR, 1234)
	require.NoError(t, err)
	require.Equal(t, int64(1142), amountMinorUSD)
}

func TestConvertToUSDReturnsExplicitErrorWhenRatesMissing(t *testing.T) {
	container := testhelpers.GetClean(t)

	_, err := container.ServiceCurrency.ConvertToUSD(container.Ctx, entities.CurrencyEUR, 1234)
	require.Error(t, err)
	require.True(t, errors.Is(err, currency.ErrRatesUnavailable))
}

func TestConvertToUSDReturnsExplicitErrorWhenRateIsZero(t *testing.T) {
	container := testhelpers.GetClean(t)
	require.NoError(t, container.Repo.SaveCurrencyRates(container.Ctx, map[entities.Currency]int64{
		entities.CurrencyUSD: currency.RateScale,
		entities.CurrencyBTC: 0,
	}, time.Now()))

	_, err := container.ServiceCurrency.ConvertToUSD(container.Ctx, entities.CurrencyBTC, 1234)
	require.Error(t, err)
	require.True(t, errors.Is(err, currency.ErrInvalidRate))
}

func TestConvertToUSDReturnsExplicitErrorWhenAmountOverflows(t *testing.T) {
	container := testhelpers.GetClean(t)
	require.NoError(t, container.Repo.SaveCurrencyRates(container.Ctx, map[entities.Currency]int64{
		entities.CurrencyUSD: currency.RateScale,
		entities.CurrencyBTC: 1,
	}, time.Now()))

	_, err := container.ServiceCurrency.ConvertToUSD(container.Ctx, entities.CurrencyBTC, math.MaxInt64)
	require.Error(t, err)
	require.True(t, errors.Is(err, currency.ErrAmountTooLarge))
}

func TestConvertToUSDFallbackDoesNotBlockTodaysRefresh(t *testing.T) {
	container := testhelpers.GetClean(t)
	require.NoError(t, container.Repo.SaveCurrencyRates(container.Ctx, map[entities.Currency]int64{
		entities.CurrencyUSD: currency.RateScale,
		entities.CurrencyEUR: 200000000,
	}, time.Now()))

	amountMinorUSD, err := container.ServiceCurrency.ConvertToUSD(container.Ctx, entities.CurrencyEUR, 1234)
	require.NoError(t, err)
	require.Equal(t, int64(617), amountMinorUSD)
}
