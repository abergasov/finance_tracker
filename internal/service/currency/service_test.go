package currency_test

import (
	"testing"

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
