package user

import (
	"context"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/logger"

	"github.com/google/uuid"
)

// UpdateUserDefaultCurrency persists the user's default currency preference.
// The currency must be a known value; callers are expected to validate first.
func (s *Service) UpdateUserDefaultCurrency(ctx context.Context, userID uuid.UUID, currency entities.Currency) error {
	if err := s.repo.UpdateUserDefaultCurrency(ctx, userID, currency.String()); err != nil {
		s.log.Error("failed update user currency", err, logger.WithUserID(userID), logger.WithString("currency", currency.String()))
		return err
	}
	return nil
}
