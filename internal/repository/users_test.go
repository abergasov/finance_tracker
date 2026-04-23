package repository_test

import (
	"finance_tracker/internal/test_helpers/seed"
	"testing"

	testhelpers "finance_tracker/internal/test_helpers"

	"github.com/stretchr/testify/require"
)

func TestSaveUserByEmailHandlesDuplicateEmailCaseInsensitively(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)

	// when
	usr := seed.NewUserBuilder().PopulateTest(t, container)

	// then
	dbUser, err := container.Repo.GetUserByID(container.Ctx, usr.ID)
	require.NoError(t, err)

	require.Equal(t, usr.Email, dbUser.Email)
	require.Equal(t, usr.Name, dbUser.Name)
}
