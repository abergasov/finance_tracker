package user

import (
	"context"

	"github.com/google/uuid"
)

// AddUserExpense creates a new category as a child of parentID and invalidates the cache.
func (s *Service) AddUserExpense(ctx context.Context, userID uuid.UUID, parentID int64, name string) (int64, error) {
	id, err := s.repo.SaveUserExpenses(ctx, userID, &parentID, name)
	if err != nil {
		return 0, err
	}
	usrExpensesCategories.Delete(userID)
	return id, nil
}

// UpdateUserExpense renames a non-root category and invalidates the cache.
func (s *Service) UpdateUserExpense(ctx context.Context, userID uuid.UUID, id int64, name string) error {
	if err := s.repo.UpdateUserExpense(ctx, userID, id, name); err != nil {
		return err
	}
	usrExpensesCategories.Delete(userID)
	return nil
}

// DeleteUserExpense deletes a non-root category (and its descendants via DB cascade)
// and invalidates the cache.
func (s *Service) DeleteUserExpense(ctx context.Context, userID uuid.UUID, id int64) error {
	if err := s.repo.DeleteUserExpense(ctx, userID, id); err != nil {
		return err
	}
	usrExpensesCategories.Delete(userID)
	return nil
}
