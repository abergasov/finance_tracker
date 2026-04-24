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
	}
	tableCategoryExpensesStr = strings.Join(tableCategoryExpensesCols, ",")
)

func (r *Repo) LoadAllUserExpenses(ctx context.Context, userID uuid.UUID) ([]*entities.UserExpensesCategoryDB, error) {
	q := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", tableCategoryExpensesStr, TableCategoryExpenses)
	return utils.QueryRowsToStruct[entities.UserExpensesCategoryDB](ctx, r.db.Client(), q, userID)
}

func (r *Repo) SaveUserExpenses(ctx context.Context, userID uuid.UUID, parent *int64, name string) (int64, error) {
	if name == "" {
		return 0, errors.New("name is required")
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
	return utils.QueryRowPrimitive[int64](ctx, r.db.Client(), q, p...)
}
