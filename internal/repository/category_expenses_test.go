package repository_test

import (
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCategoryExpenses(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	user := seed.NewUserBuilder().PopulateTest(t, container)
	expensesList, err := container.Repo.LoadAllUserExpenses(container.Ctx, user.ID)
	require.NoError(t, err)
	require.Empty(t, expensesList)

	parentID, err := container.Repo.SaveUserExpenses(container.Ctx, user.ID, nil, uuid.NewString())
	require.NoError(t, err)
	_, err = container.Repo.SaveUserExpenses(container.Ctx, user.ID, new(parentID), uuid.NewString())
	require.NoError(t, err)

	// when
	expensesList, err = container.Repo.LoadAllUserExpenses(container.Ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, expensesList, 2)
}
