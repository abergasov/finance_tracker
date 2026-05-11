package repository_test

import (
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpenses(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	usr := seed.NewUserBuilder().PopulateTest(t, container.Repo)
	require.NoError(t, container.Repo.EnsureRootCategories(container.Ctx, usr.ID))
	rows, err := container.Repo.LoadAllUserExpenses(container.Ctx, usr.ID)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	// when
	require.NoError(t, container.Repo.SaveExpenseRecord(container.Ctx, usr.ID, rows[0].ID, "USD", 1, 1))
	require.NoError(t, container.Repo.SaveExpenseRecord(container.Ctx, usr.ID, rows[0].ID, "USD", 2, 2))
	require.NoError(t, container.Repo.SaveExpenseRecord(container.Ctx, usr.ID, rows[0].ID, "USD", 3, 3))

	// then
	records, err := container.Repo.LoadExpenseRecords(container.Ctx, usr.ID)
	require.NoError(t, err)
	require.Len(t, records, 3)
	require.Equal(t, int64(1), records[0].AmountMinorUSD)
	require.Equal(t, int64(2), records[1].AmountMinorUSD)
	require.Equal(t, int64(3), records[2].AmountMinorUSD)
}
