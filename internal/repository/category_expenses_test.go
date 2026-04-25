package repository_test

import (
	"database/sql"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/repository"
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
	user := seed.NewUserBuilder().PopulateTest(t, container)
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
	usr := seed.NewUserBuilder().PopulateTest(t, container)

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
}

func TestCategoryRootUniquenessMigrationRepairsLegacyDuplicateRoots(t *testing.T) {
	container := testhelpers.GetClean(t)
	usr := seed.NewUserBuilder().PopulateTest(t, container)

	tx, err := container.DB.Client().BeginTxx(container.Ctx, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tx.Rollback())
	})

	_, err = tx.ExecContext(container.Ctx, `DROP INDEX IF EXISTS category_expenses_user_root_uidx`)
	require.NoError(t, err)

	insertCategory := func(parent sql.NullInt64, name string) int64 {
		var id int64
		err := tx.QueryRowContext(container.Ctx, `
			INSERT INTO category_expenses (user_id, parent_id, name)
			VALUES ($1, $2, $3)
			RETURNING id
		`, usr.ID, parent, name).Scan(&id)
		require.NoError(t, err)
		return id
	}

	mandatoryKeep := insertCategory(sql.NullInt64{}, "mandatory")
	mandatoryDuplicate := insertCategory(sql.NullInt64{}, "mandatory")
	mandatoryChild := insertCategory(sql.NullInt64{Int64: mandatoryDuplicate, Valid: true}, "rent")
	optionalKeep := insertCategory(sql.NullInt64{}, "optional")
	optionalDuplicate := insertCategory(sql.NullInt64{}, "optional")
	optionalChild := insertCategory(sql.NullInt64{Int64: optionalDuplicate, Valid: true}, "travel")

	legacyRows, err := utils.QueryRowsToStruct[entities.UserExpensesCategoryDB](container.Ctx, tx,
		`SELECT id, user_id, parent_id, name FROM category_expenses WHERE user_id = $1 ORDER BY id`,
		usr.ID,
	)
	require.NoError(t, err)
	require.Len(t, legacyRows, 6)

	_, err = tx.ExecContext(container.Ctx, `
		UPDATE category_expenses AS child
		SET parent_id = canon.canonical_id
		FROM (
			SELECT
				dup.id            AS dup_id,
				grp.canonical_id
			FROM category_expenses AS dup
			JOIN (
				SELECT user_id, name, MIN(id) AS canonical_id
				FROM category_expenses
				WHERE parent_id IS NULL
				GROUP BY user_id, name
				HAVING COUNT(*) > 1
			) AS grp ON dup.user_id = grp.user_id AND dup.name = grp.name
			WHERE dup.parent_id IS NULL
			AND dup.id > grp.canonical_id
		) AS canon
		WHERE child.parent_id = canon.dup_id
	`)
	require.NoError(t, err)

	_, err = tx.ExecContext(container.Ctx, `
		DELETE FROM category_expenses
		WHERE parent_id IS NULL
		AND id NOT IN (
			SELECT MIN(id)
			FROM category_expenses
			WHERE parent_id IS NULL
			GROUP BY user_id, name
		)
	`)
	require.NoError(t, err)

	_, err = tx.ExecContext(container.Ctx, `
		CREATE UNIQUE INDEX category_expenses_user_root_uidx
		ON category_expenses (user_id, name)
		WHERE parent_id IS NULL
	`)
	require.NoError(t, err)

	rows, err := utils.QueryRowsToStruct[entities.UserExpensesCategoryDB](container.Ctx, tx,
		`SELECT id, user_id, parent_id, name FROM category_expenses WHERE user_id = $1 ORDER BY id`,
		usr.ID,
	)
	require.NoError(t, err)
	require.Len(t, rows, 4)

	rootIDs := map[string]int64{}
	for _, row := range rows {
		switch row.Name {
		case "mandatory":
			require.False(t, row.ParentID.Valid)
			rootIDs[row.Name] = row.ID
		case "optional":
			require.False(t, row.ParentID.Valid)
			rootIDs[row.Name] = row.ID
		case "rent":
			require.Equal(t, mandatoryChild, row.ID)
			require.True(t, row.ParentID.Valid)
			require.Equal(t, mandatoryKeep, row.ParentID.Int64)
		case "travel":
			require.Equal(t, optionalChild, row.ID)
			require.True(t, row.ParentID.Valid)
			require.Equal(t, optionalKeep, row.ParentID.Int64)
		default:
			t.Fatalf("unexpected category row: %#v", row)
		}
	}

	require.Equal(t, mandatoryKeep, rootIDs["mandatory"])
	require.Equal(t, optionalKeep, rootIDs["optional"])
	require.Contains(t, rootIDs, "mandatory")
	require.Contains(t, rootIDs, "optional")

	var indexName sql.NullString
	require.NoError(t, tx.QueryRowContext(container.Ctx, `SELECT to_regclass('public.category_expenses_user_root_uidx')`).Scan(&indexName))
	require.True(t, indexName.Valid)
	require.Equal(t, "category_expenses_user_root_uidx", indexName.String)
}

func TestUpdateAndDeleteUserExpenseRejectRootCategories(t *testing.T) {
	container := testhelpers.GetClean(t)
	usr := seed.NewUserBuilder().PopulateTest(t, container)

	require.NoError(t, container.Repo.EnsureRootCategories(container.Ctx, usr.ID))
	roots := rootCategoryIDs(t, mustLoadCategories(t, container, usr.ID))

	err := container.Repo.UpdateUserExpense(container.Ctx, usr.ID, roots["mandatory"], uuid.NewString())
	require.ErrorIs(t, err, repository.ErrProtectedRoot)

	err = container.Repo.DeleteUserExpense(container.Ctx, usr.ID, roots["optional"])
	require.ErrorIs(t, err, repository.ErrProtectedRoot)

	rows := mustLoadCategories(t, container, usr.ID)
	require.Len(t, rows, 2)
	require.Equal(t, roots, rootCategoryIDs(t, rows))
}

func TestDeleteUserExpenseCascadesDescendants(t *testing.T) {
	container := testhelpers.GetClean(t)
	usr := seed.NewUserBuilder().PopulateTest(t, container)

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
	require.ElementsMatch(t, []string{"mandatory", "optional", "travel"}, categoryNames(rows))

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

func categoryNames(rows []*entities.UserExpensesCategoryDB) []string {
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}
	return names
}
