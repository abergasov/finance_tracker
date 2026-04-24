package repository_test

import (
	"finance_tracker/internal/entities"
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"
	"finance_tracker/internal/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCategoryExpenses(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	user := seed.NewUserBuilder().PopulateTest(t, container.Repo)
	expensesList, err := container.Repo.LoadAllUserExpenses(container.Ctx, user.ID)
	require.NoError(t, err)
	require.Empty(t, expensesList)

	parentID, err := container.Repo.SaveUserExpenses(container.Ctx, user.ID, nil, uuid.NewString())
	require.NoError(t, err)
	_, err = container.Repo.SaveUserExpenses(container.Ctx, user.ID, &parentID, uuid.NewString())
	require.NoError(t, err)

	// when
	expensesList, err = container.Repo.LoadAllUserExpenses(container.Ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, expensesList, 2)
}

func TestEnsureRootCategoriesIsIdempotent(t *testing.T) {
	container := testhelpers.GetClean(t)
	usr := seed.NewUserBuilder().PopulateTest(t, container.Repo)

	var firstRoots map[string]int64
	for i := 0; i < 3; i++ {
		require.NoError(t, container.Repo.EnsureRootCategories(container.Ctx, usr.ID))

		rows, err := container.Repo.LoadAllUserExpenses(container.Ctx, usr.ID)
		require.NoError(t, err)
		require.Len(t, rows, 2)

		roots := rootCategoryIDs(t, rows)
		if i == 0 {
			firstRoots = roots
			continue
		}
		require.Equal(t, firstRoots, roots)
	}

	t.Run("delete", func(t *testing.T) {
		roots := rootCategoryIDs(t, mustLoadCategories(t, container, usr.ID))

		// when
		require.NoError(t, container.Repo.UpdateUserExpense(container.Ctx, usr.ID, roots["mandatory"], uuid.NewString()))
		require.NoError(t, container.Repo.DeleteUserExpense(container.Ctx, usr.ID, roots["optional"]))

		// then
		rows := mustLoadCategories(t, container, usr.ID)
		require.Len(t, rows, 2)
	})
}

func TestDeleteUserExpenseCascadesDescendants(t *testing.T) {
	container := testhelpers.GetClean(t)
	usr := seed.NewUserBuilder().PopulateTest(t, container.Repo)

	require.NoError(t, container.Repo.EnsureRootCategories(container.Ctx, usr.ID))
	roots := rootCategoryIDs(t, mustLoadCategories(t, container, usr.ID))

	mandatoryParentID := roots["mandatory"]
	optionalParentID := roots["optional"]

	branchID, err := container.Repo.SaveUserExpenses(container.Ctx, usr.ID, &mandatoryParentID, "household")
	require.NoError(t, err)
	childID, err := container.Repo.SaveUserExpenses(container.Ctx, usr.ID, &branchID, "utilities")
	require.NoError(t, err)
	_, err = container.Repo.SaveUserExpenses(container.Ctx, usr.ID, &optionalParentID, "travel")
	require.NoError(t, err)

	require.NoError(t, container.Repo.DeleteUserExpense(container.Ctx, usr.ID, branchID))

	rows := mustLoadCategories(t, container, usr.ID)
	require.Len(t, rows, 3)
	require.ElementsMatch(t, []string{"mandatory", "optional", "travel"}, utils.StringsFromObjectSlice(rows, func(db *entities.UserExpensesCategoryDB) string {
		return db.Name
	}))

	for _, row := range rows {
		require.NotEqual(t, branchID, row.ID)
		require.NotEqual(t, childID, row.ID)
	}
}

func mustLoadCategories(t *testing.T, container *testhelpers.TestContainer, userID uuid.UUID) []*entities.UserExpensesCategoryDB {
	t.Helper()

	rows, err := container.Repo.LoadAllUserExpenses(container.Ctx, userID)
	require.NoError(t, err)
	return rows
}

func rootCategoryIDs(t *testing.T, rows []*entities.UserExpensesCategoryDB) map[string]int64 {
	t.Helper()

	roots := make(map[string]int64, 2)
	for _, row := range rows {
		require.False(t, row.ParentID.Valid)
		roots[row.Name] = row.ID
	}
	require.Len(t, roots, 2)
	return roots
}
