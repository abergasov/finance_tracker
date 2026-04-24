package user

import (
	"context"
	"finance_tracker/internal/entities"

	"github.com/google/uuid"
)

func (s *Service) AddUserExpense(ctx context.Context, userID uuid.UUID, exp *entities.UserExpensesCategoryDB) error {
	_, err := s.repo.SaveUserExpenses(ctx, userID, new(exp.ParentID.Int64), exp.Name)
	if err != nil {
		return err
	}
	usrExpensesCategories.Delete(userID)
	return nil
}
