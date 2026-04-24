package repository

import (
	"context"
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

func (r *Repo) SaveUserExpenses(ctx context.Context, expenses *entities.UserExpensesCategoryDB) (int64, error) {
	q, p := utils.GenerateInsertSQL(TableCategoryExpenses, map[string]any{
		"user_id":   expenses.UserID,
		"parent_id": expenses.ParentID,
		"name":      expenses.Name,
	})
	q += " RETURNING id"
	return utils.QueryRowPrimitive[int64](ctx, r.db.Client(), q, p...)
}
