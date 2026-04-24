package repository_test

import (
	"testing"
	"time"

	"finance_tracker/internal/utils"

	"finance_tracker/internal/entities"
	testhelpers "finance_tracker/internal/test_helpers"

	"github.com/stretchr/testify/require"
)

func TestCurrencies(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	rates, err := container.Repo.LoadAllCurrencies(container.Ctx)
	require.NoError(t, err)
	require.Empty(t, rates)

	// when
	require.NoError(t, container.Repo.SaveCurrencyRates(container.Ctx, map[entities.Currency]int64{
		entities.CurrencyAED: 1,
		entities.CurrencyUSD: 1,
		entities.CurrencyRUB: 123,
	}, time.Now()))

	// then
	rates, err = container.Repo.LoadAllCurrencies(container.Ctx)
	require.NoError(t, err)
	require.Len(t, rates, 1)
	require.Equal(t, utils.TimeToDayIntNum(time.Now()), rates[0].TimestampNum)
	require.Equal(t, int64(1), rates[0].CurrencyData[entities.CurrencyAED.String()])
	require.Equal(t, int64(1), rates[0].CurrencyData[entities.CurrencyUSD.String()])
	require.Equal(t, int64(123), rates[0].CurrencyData[entities.CurrencyRUB.String()])
}
