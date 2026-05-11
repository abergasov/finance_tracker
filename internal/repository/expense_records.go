package repository

import (
	"context"
	"finance_tracker/internal/entities"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"finance_tracker/internal/utils"
)

const TableExpenseRecords = "expense_records"

var (
	tableExpenseRecordsCols = []string{
		"user_id",
		"category_id",
		"currency",
		"amount_minor",
		"created_at",
	}
	tableExpenseRecordsColsStr = strings.Join(tableExpenseRecordsCols, ",")
)

// SaveExpenseRecord inserts a new expense record for the given user.
// amountMinor is the amount in currency minor units (e.g. cents).
func (r *Repo) SaveExpenseRecord(ctx context.Context, userID uuid.UUID, categoryID int64, currency string, amountMinor int64) error {
	q, p := utils.GenerateInsertSQL(TableExpenseRecords, map[string]any{
		"user_id":      userID,
		"category_id":  categoryID,
		"currency":     currency,
		"amount_minor": amountMinor,
		"created_at":   time.Now(),
	})
	if _, err := r.db.Client().ExecContext(ctx, q, p...); err != nil {
		return fmt.Errorf("save expense record: %w", err)
	}
	return nil
}

func (r *Repo) LoadExpenseRecords(ctx context.Context, userID uuid.UUID) ([]*entities.ExpenseRecord, error) {
	q := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", tableExpenseRecordsColsStr, TableExpenseRecords)
	return utils.QueryRowsToStruct[entities.ExpenseRecord](ctx, r.db.Client(), q, userID)
}
