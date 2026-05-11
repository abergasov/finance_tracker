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
		"color",
	}
	tableCategoryExpensesStr = strings.Join(tableCategoryExpensesCols, ",")
)

func (r *Repo) LoadAllUserExpenses(ctx context.Context, userID uuid.UUID) ([]*entities.UserExpensesCategoryDB, error) {
	q := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", tableCategoryExpensesStr, TableCategoryExpenses)
	return utils.QueryRowsToStruct[entities.UserExpensesCategoryDB](ctx, r.db.Client(), q, userID)
}

func (r *Repo) LoadUserExpense(ctx context.Context, userID uuid.UUID, expenseID int64) (*entities.UserExpensesCategoryDB, error) {
	q := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1 AND user_id = $2", tableCategoryExpensesStr, TableCategoryExpenses)
	return utils.QueryRowToStruct[entities.UserExpensesCategoryDB](ctx, r.db.Client(), q, expenseID, userID)
}

// EnsureRootCategories creates the fixed mandatory/optional root categories for a user if they do not already exist.
// Safe to call multiple times and race-safe: it relies on
// the partial unique index category_expenses_user_root_uidx (user_id, name WHERE parent_id
// IS NULL) to make the insert atomic, avoiding the TOCTOU window of the previous
// SELECT+INSERT pattern.
func (r *Repo) EnsureRootCategories(ctx context.Context, userID uuid.UUID) error {
	for _, name := range []string{"mandatory", "optional"} {
		color := utils.DeriveHexColor(name)
		q := fmt.Sprintf(`
			INSERT INTO %s (user_id, name, color)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, name) WHERE parent_id IS NULL DO NOTHING`,
			TableCategoryExpenses)
		if _, err := r.db.Client().ExecContext(ctx, q, userID, name, color); err != nil {
			return fmt.Errorf("ensure root category %q: %w", name, err)
		}
	}
	return nil
}

func (r *Repo) SaveUserExpenses(ctx context.Context, userID uuid.UUID, parent *int64, name, color string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name is required")
	}
	normalizedColor, ok := utils.NormalizeHexColor(color)
	if !ok {
		return 0, errors.New("color is invalid")
	}

	// Verify the parent belongs to this user before inserting.
	if parent != nil {
		if expense, err := r.LoadUserExpense(ctx, userID, *parent); err != nil || expense == nil {
			return 0, fmt.Errorf("invalid parent expense: %w", err)
		}
	}
	q, p := utils.GenerateInsertSQL(TableCategoryExpenses, map[string]any{
		"user_id": userID,
		"parent_id": sql.NullInt64{
			Int64: utils.FromPointer(parent),
			Valid: parent != nil,
		},
		"name":  name,
		"color": normalizedColor,
	})
	q += " RETURNING id"
	id, err := utils.QueryRowPrimitive[int64](ctx, r.db.Client(), q, p...)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, errors.New("a category with this name already exists under this parent")
		}
		return 0, err
	}
	return id, nil
}

// UpdateUserExpense renames a category
func (r *Repo) UpdateUserExpense(ctx context.Context, userID uuid.UUID, id int64, name, color string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name is required")
	}
	normalizedColor, ok := utils.NormalizeHexColor(color)
	if !ok {
		return errors.New("color is invalid")
	}
	q := fmt.Sprintf("UPDATE %s SET name = $1, color = $2, updated_at = now() WHERE id = $3 AND user_id = $4 AND parent_id IS NOT NULL", TableCategoryExpenses)
	_, err := r.db.Client().ExecContext(ctx, q, name, normalizedColor, id, userID)
	return err
}

// DeleteUserExpense deletes a category and all its descendants (via DB cascade).
func (r *Repo) DeleteUserExpense(ctx context.Context, userID uuid.UUID, id int64) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND user_id = $2 AND parent_id IS NOT NULL", TableCategoryExpenses)
	_, err := r.db.Client().ExecContext(ctx, q, id, userID)
	return err
}
