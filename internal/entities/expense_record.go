package entities

import (
	"time"

	"github.com/google/uuid"
)

// ExpenseRecord is a single persisted expense entry.
type ExpenseRecord struct {
	ID          int64     `db:"id"           json:"id"`
	UserID      uuid.UUID `db:"user_id"      json:"user_id"`
	CategoryID  int64     `db:"category_id"  json:"category_id"`
	Currency    string    `db:"currency"     json:"currency"`
	AmountMinor int64     `db:"amount_minor" json:"amount_minor"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}
