package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"finance_tracker/internal/entities"
	"finance_tracker/internal/logger"
)

// ErrInvalidCategory is returned when the requested category does not exist
// or does not belong to the authenticated user.
var ErrInvalidCategory = errors.New("invalid category")

// ErrRootCategory is returned when the caller attempts to record an expense
// against a root category (mandatory / optional).
var ErrRootCategory = errors.New("category must not be a root category")

// CreateExpenseRecord validates and persists an expense record for the authenticated user.
// amountMinor must already be a positive integer in currency minor units (use
// ParseAmountToMinor to convert the user-provided string before calling here).
// Validation:
//   - categoryID must identify a category owned by userID
//   - that category must not be a root category (parent_id must not be NULL)
//
// The currency amount is converted to USD minor units via the currency service
// and stored alongside the original currency/amount.
func (s *Service) CreateExpenseRecord(ctx context.Context, userID uuid.UUID, categoryID int64, currency entities.Currency, amountMinor int64) error {
	category, err := s.repo.LoadUserExpense(ctx, userID, categoryID)
	if err != nil || category == nil {
		return ErrInvalidCategory
	}
	if !category.ParentID.Valid {
		return ErrRootCategory
	}

	amountMinorUSD := amountMinor
	if currency != entities.CurrencyUSD {
		amountMinorUSD, err = s.currencySvc.ConvertToUSD(ctx, currency, amountMinor)
		if err != nil {
			return fmt.Errorf("convert to USD: %w", err)
		}
	}

	if err = s.repo.SaveExpenseRecord(ctx, userID, categoryID, currency.String(), amountMinor, amountMinorUSD); err != nil {
		s.log.Error("failed to save expense record", err,
			logger.WithUserID(userID),
			logger.WithInt64("category_id", categoryID),
			logger.WithString("currency", currency.String()),
			logger.WithInt64("amount_minor", amountMinor),
		)
		return err
	}
	return nil
}
