package user

import (
	"context"
	"finance_tracker/internal/logger"
	"finance_tracker/internal/utils"
	"fmt"

	"github.com/google/uuid"
)

// AddUserExpense creates a new category as a child of parentID and invalidates the cache.
func (s *Service) AddUserExpense(ctx context.Context, userID uuid.UUID, parentID int64, name, color string) (int64, error) {
	normalizedColor, ok := utils.NormalizeHexColor(color)
	if !ok {
		return 0, fmt.Errorf("invalid color: %s", color)
	}
	id, err := s.repo.SaveUserExpenses(ctx, userID, &parentID, name, normalizedColor)
	if err != nil {
		s.log.Error("failed to add user expense category", err,
			logger.WithUserID(userID),
			logger.WithString("category_name", name),
			logger.WithString("category_color", normalizedColor),
			logger.WithInt64("parent_id", parentID),
		)
		return 0, err
	}
	usrExpensesCategories.Delete(userID)
	return id, nil
}

// UpdateUserExpense renames a non-root category and invalidates the cache.
func (s *Service) UpdateUserExpense(ctx context.Context, userID uuid.UUID, id int64, name, color string) error {
	normalizedColor, ok := utils.NormalizeHexColor(color)
	if !ok {
		return fmt.Errorf("invalid color: %s", color)
	}
	if err := s.repo.UpdateUserExpense(ctx, userID, id, name, normalizedColor); err != nil {
		s.log.Error("failed to update user expense category", err,
			logger.WithUserID(userID),
			logger.WithString("category_name", name),
			logger.WithString("category_color", normalizedColor),
		)
		return err
	}
	usrExpensesCategories.Delete(userID)
	return nil
}

// DeleteUserExpense deletes a non-root category (and its descendants via DB cascade)
// and invalidates the cache.
func (s *Service) DeleteUserExpense(ctx context.Context, userID uuid.UUID, id int64) error {
	if err := s.repo.DeleteUserExpense(ctx, userID, id); err != nil {
		s.log.Error("failed to delete user expense category", err,
			logger.WithUserID(userID),
			logger.WithInt64("category_id", id),
		)
		return err
	}
	usrExpensesCategories.Delete(userID)
	return nil
}
