package entities

import (
	"database/sql"
	"time"

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
	ID       int64                   `db:"id" json:"id"`
	Children []*UserExpensesCategory `json:"children,omitempty"`
	Name     string                  `json:"name"`
}

func (e *UserExpensesCategory) FindCategoryByID(id int64) *UserExpensesCategory {
	if e == nil {
		return nil
	}
	if e.ID == id {
		return e
	}
	for _, child := range e.Children {
		if found := child.FindCategoryByID(id); found != nil {
			return found
		}
	}
	return nil
}

type UserExpensesCategoryDB struct {
	ID        int64         `db:"id" json:"id"`
	UserID    uuid.UUID     `db:"user_id" json:"userId"`
	ParentID  sql.NullInt64 `db:"parent_id" json:"parentId"`
	Name      string        `db:"name" json:"name"`
	CreatedAt time.Time     `db:"created_at" json:"-"`
	UpdatedAt time.Time     `db:"updated_at" json:"-"`
}
