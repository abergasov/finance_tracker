package seed

import (
	"context"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type UserBuilder struct {
	user *entities.User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: &entities.User{
			Email: uuid.NewString(),
			Name:  uuid.NewString(),
		},
	}
}

func (d *UserBuilder) Build() *entities.User {
	return d.user
}

func (d *UserBuilder) PopulateTest(t testing.TB, repo *repository.Repo) *entities.User {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	usr := d.Build()
	dbUsr, err := repo.SaveUserByEmail(ctx, usr.Email, usr.Name)
	require.NoError(t, err)
	return dbUsr
}
