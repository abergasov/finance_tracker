package user

import (
	"context"
	"finance_tracker/internal/logger"

	"github.com/google/uuid"
)

// AddUserExpense creates a new category as a child of parentID and invalidates the cache.
func (s *Service) AddUserExpense(ctx context.Context, userID uuid.UUID, parentID int64, name string) (int64, error) {
	id, err := s.repo.SaveUserExpenses(ctx, userID, &parentID, name)
	if err != nil {
		s.log.Error("failed to add user expense category", err,
			logger.WithUserID(userID),
			logger.WithString("category_name", name),
			logger.WithInt64("parent_id", parentID),
		)
		return 0, err
	}
	usrExpensesCategories.Delete(userID)
	return id, nil
}

// UpdateUserExpense renames a non-root category and invalidates the cache.
func (s *Service) UpdateUserExpense(ctx context.Context, userID uuid.UUID, id int64, name string) error {
	if err := s.repo.UpdateUserExpense(ctx, userID, id, name); err != nil {
		s.log.Error("failed to update user expense category", err,
			logger.WithUserID(userID),
			logger.WithString("category_name", name),
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
