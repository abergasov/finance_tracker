package repository

import (
	"context"
	"database/sql"
	"errors"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/utils"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Domain-level sentinel errors for category operations.
// Callers should use errors.Is to match these.
var (
	// ErrParentNotFound is returned when the requested parent category does not exist
	// or does not belong to the authenticated user.
	ErrParentNotFound = errors.New("parent category not found")

	// ErrCategoryNotFound is returned when the target category does not exist
	// or does not belong to the authenticated user.
	ErrCategoryNotFound = errors.New("category not found")

	// ErrProtectedRoot is returned when an attempt is made to mutate a root category
	// (mandatory / optional) which must remain fixed.
	ErrProtectedRoot = errors.New("root category cannot be modified")

	// ErrDuplicateName is returned when the requested name already exists under
	// the same parent for the same user (unique constraint violation).
	ErrDuplicateName = errors.New("a category with this name already exists under this parent")
)

const (
	TableCategoryExpenses = "category_expenses"
)

var (
	tableCategoryExpensesCols = []string{
		"id",
		"user_id",
		"parent_id",
		"name",
	}
	tableCategoryExpensesStr = strings.Join(tableCategoryExpensesCols, ",")
)

func (r *Repo) LoadAllUserExpenses(ctx context.Context, userID uuid.UUID) ([]*entities.UserExpensesCategoryDB, error) {
	q := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", tableCategoryExpensesStr, TableCategoryExpenses)
	return utils.QueryRowsToStruct[entities.UserExpensesCategoryDB](ctx, r.db.Client(), q, userID)
}

// EnsureRootCategories creates the fixed mandatory/optional root categories for a user
// if they do not already exist. Safe to call multiple times and race-safe: it relies on
// the partial unique index category_expenses_user_root_uidx (user_id, name WHERE parent_id
// IS NULL) to make the insert atomic, avoiding the TOCTOU window of the previous
// SELECT+INSERT pattern.
func (r *Repo) EnsureRootCategories(ctx context.Context, userID uuid.UUID) error {
	for _, name := range []string{"mandatory", "optional"} {
		q := fmt.Sprintf(`
			INSERT INTO %s (user_id, name)
			VALUES ($1, $2)
			ON CONFLICT (user_id, name) WHERE parent_id IS NULL DO NOTHING`,
			TableCategoryExpenses)
		if _, err := r.db.Client().ExecContext(ctx, q, userID, name); err != nil {
			return fmt.Errorf("ensure root category %q: %w", name, err)
		}
	}
	return nil
}

func (r *Repo) SaveUserExpenses(ctx context.Context, userID uuid.UUID, parent *int64, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name is required")
	}
	// Verify the parent belongs to this user before inserting.
	if parent != nil {
		if err := r.checkParentOwnership(ctx, userID, *parent); err != nil {
			return 0, err
		}
	}
	q, p := utils.GenerateInsertSQL(TableCategoryExpenses, map[string]any{
		"user_id": userID,
		"parent_id": sql.NullInt64{
			Int64: utils.FromPointer(parent),
			Valid: parent != nil,
		},
		"name": name,
	})
	q += " RETURNING id"
	id, err := utils.QueryRowPrimitive[int64](ctx, r.db.Client(), q, p...)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrDuplicateName
		}
		return 0, err
	}
	return id, nil
}

// checkParentOwnership returns ErrParentNotFound if the parent row does not
// exist or does not belong to userID.
func (r *Repo) checkParentOwnership(ctx context.Context, userID uuid.UUID, parentID int64) error {
	q := fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND user_id = $2)",
		TableCategoryExpenses,
	)
	var exists bool
	if err := r.db.Client().QueryRowContext(ctx, q, parentID, userID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrParentNotFound
	}
	return nil
}

// checkCategoryOwnershipAndRoot returns ErrCategoryNotFound when the category
// does not exist or belongs to another user, and ErrProtectedRoot when the
// category is a root (parent_id IS NULL).
func (r *Repo) checkCategoryOwnershipAndRoot(ctx context.Context, userID uuid.UUID, id int64) error {
	q := fmt.Sprintf(
		"SELECT parent_id IS NULL FROM %s WHERE id = $1 AND user_id = $2",
		TableCategoryExpenses,
	)
	var isRoot bool
	err := r.db.Client().QueryRowContext(ctx, q, id, userID).Scan(&isRoot)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCategoryNotFound
	}
	if err != nil {
		return err
	}
	if isRoot {
		return ErrProtectedRoot
	}
	return nil
}

// isUniqueViolation reports whether err is a PostgreSQL unique-constraint
// violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// UpdateUserExpense renames a category. Returns ErrCategoryNotFound if the
// category does not exist or belongs to another user, ErrProtectedRoot if it
// is a root category, ErrDuplicateName if the name already exists under the
// same parent, or a storage error for unexpected failures.
func (r *Repo) UpdateUserExpense(ctx context.Context, userID uuid.UUID, id int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name is required")
	}
	if err := r.checkCategoryOwnershipAndRoot(ctx, userID, id); err != nil {
		return err
	}
	q := fmt.Sprintf(
		"UPDATE %s SET name = $1, updated_at = now() WHERE id = $2 AND user_id = $3 AND parent_id IS NOT NULL",
		TableCategoryExpenses,
	)
	_, err := r.db.Client().ExecContext(ctx, q, name, id, userID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateName
		}
		return err
	}
	return nil
}

// DeleteUserExpense deletes a category and all its descendants (via DB cascade).
// Returns ErrCategoryNotFound if the category does not exist or belongs to
// another user, ErrProtectedRoot if it is a root category, or a storage error
// for unexpected failures.
func (r *Repo) DeleteUserExpense(ctx context.Context, userID uuid.UUID, id int64) error {
	if err := r.checkCategoryOwnershipAndRoot(ctx, userID, id); err != nil {
		return err
	}
	q := fmt.Sprintf(
		"DELETE FROM %s WHERE id = $1 AND user_id = $2 AND parent_id IS NOT NULL",
		TableCategoryExpenses,
	)
	_, err := r.db.Client().ExecContext(ctx, q, id, userID)
	return err
}
