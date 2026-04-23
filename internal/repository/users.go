package repository

import (
	"context"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/utils"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	TableUsers = "users"
)

var (
	tableUsersCols = []string{
		"u_id",
		"email",
		"user_name",
	}
	tableUsersColsStr = strings.Join(tableUsersCols, ",")
)

func (r *Repo) SaveUserByEmail(ctx context.Context, email, name string) (*entities.User, error) {
	userID := uuid.New()
	q, p := utils.GenerateInsertSQL(TableUsers, map[string]any{
		"u_id":       userID,
		"created_at": time.Now(),
		"updated_at": time.Now(),
		"email":      email,
		"user_name":  name,
	})
	q += " ON CONFLICT DO NOTHING"
	if _, err := r.db.Client().ExecContext(ctx, q, p...); err != nil {
		return nil, fmt.Errorf("google account email is missing or not verified")
	}
	return r.GetUserByID(ctx, userID)
}

func (r *Repo) GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	q := fmt.Sprintf("SELECT %s FROM %s WHERE u_id = $1", tableUsersColsStr, TableUsers)
	return utils.QueryRowToStruct[entities.User](ctx, r.db.Client(), q, id)
}
