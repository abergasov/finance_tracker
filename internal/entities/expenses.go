package entities

import (
	"database/sql"

	"github.com/google/uuid"
)

// UserExpenses provide list of user configured expenses
// - mandatory
// -- rent
// -- taxes
// -- insurance
// - optional
// -- travel
// --- food
// --- hotel
// -- restaurants
// -- bars
type UserExpenses struct {
	MandatoryExpenses UserExpensesCategory `db:"mandatory_expenses" json:"mandatoryExpenses"`
	OptionalExpenses  UserExpensesCategory `db:"optional_expenses" json:"optionalExpenses"`
}

type UserExpensesCategory struct {
	ID           int64                 `db:"id" json:"id"`
	ChildExpense *UserExpensesCategory `json:"childExpense"`
	Name         string                `json:"name"`
}

type UserExpensesCategoryDB struct {
	ID       int64         `db:"id" json:"id"`
	UserID   uuid.UUID     `db:"user_id" json:"userId"`
	ParentID sql.NullInt64 `db:"parent_id" json:"parentId"`
	Name     string        `db:"name" json:"name"`
}
