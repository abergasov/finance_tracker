package user_test

import (
	"database/sql"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/service/user"
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"
	"finance_tracker/internal/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildUserExpensesTree(t *testing.T) {
	tests := map[string]struct {
		rows    []*entities.UserExpensesCategoryDB
		want    *entities.UserExpenses
		wantErr string
	}{
		"success": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "mandatory"},
				{ID: 2, Name: "rent", ParentID: sql.NullInt64{Int64: 1, Valid: true}},
				{ID: 3, Name: "optional"},
				{ID: 4, Name: "travel", ParentID: sql.NullInt64{Int64: 3, Valid: true}},
				{ID: 5, Name: "hotel", ParentID: sql.NullInt64{Int64: 4, Valid: true}},
			},
			want: &entities.UserExpenses{
				MandatoryExpenses: entities.UserExpensesCategory{
					ID:   1,
					Name: "mandatory",
					Children: []*entities.UserExpensesCategory{
						{ID: 2, Name: "rent"},
					},
				},
				OptionalExpenses: entities.UserExpensesCategory{
					ID:   3,
					Name: "optional",
					Children: []*entities.UserExpensesCategory{
						{
							ID:   4,
							Name: "travel",
							Children: []*entities.UserExpensesCategory{
								{ID: 5, Name: "hotel"},
							},
						},
					},
				},
			},
		},
		"multiple children": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "mandatory"},
				{ID: 2, Name: "optional"},
				{ID: 3, Name: "rent", ParentID: sql.NullInt64{Int64: 1, Valid: true}},
				{ID: 4, Name: "taxes", ParentID: sql.NullInt64{Int64: 1, Valid: true}},
			},
			want: &entities.UserExpenses{
				MandatoryExpenses: entities.UserExpensesCategory{
					ID:   1,
					Name: "mandatory",
					Children: []*entities.UserExpensesCategory{
						{ID: 3, Name: "rent"},
						{ID: 4, Name: "taxes"},
					},
				},
				OptionalExpenses: entities.UserExpensesCategory{
					ID:   2,
					Name: "optional",
				},
			},
		},
		"nil row ignored": {
			rows: []*entities.UserExpensesCategoryDB{
				nil,
				{ID: 1, Name: "mandatory"},
				{ID: 2, Name: "optional"},
			},
			want: &entities.UserExpenses{
				MandatoryExpenses: entities.UserExpensesCategory{
					ID:   1,
					Name: "mandatory",
				},
				OptionalExpenses: entities.UserExpensesCategory{
					ID:   2,
					Name: "optional",
				},
			},
		},
		"duplicate id": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "mandatory"},
				{ID: 1, Name: "optional"},
			},
			wantErr: "duplicate expense category id: 1",
		},
		"duplicate mandatory root": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "mandatory"},
				{ID: 2, Name: "mandatory"},
				{ID: 3, Name: "optional"},
			},
			wantErr: "duplicate root category: mandatory",
		},
		"duplicate optional root": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "mandatory"},
				{ID: 2, Name: "optional"},
				{ID: 3, Name: "optional"},
			},
			wantErr: "duplicate root category: optional",
		},
		"missing parent": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "mandatory"},
				{ID: 2, Name: "rent", ParentID: sql.NullInt64{Int64: 404, Valid: true}},
			},
			wantErr: "missing parent category: id=2 parent_id=404",
		},
		"unknown root": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "bad"},
			},
			wantErr: "unknown root expense category: bad",
		},
		"empty rows": {
			rows: []*entities.UserExpensesCategoryDB{},
			want: &entities.UserExpenses{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := user.BuildUserExpensesTree(tt.rows)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServeHomePageSeedsRootCategories(t *testing.T) {
	container := testhelpers.GetClean(t)
	usr := seed.NewUserBuilder().PopulateTest(t, container)

	token, err := utils.SignClaims(&entities.SignedTokenClaims{
		Kind:   utils.AuthTokenKind,
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
	require.NoError(t, err)

	homeData, err := container.ServiceUser.ServeHomePage(container.Ctx, token)
	require.NoError(t, err)
	require.Equal(t, "mandatory", homeData.UserExpensesCategories.MandatoryExpenses.Name)
	require.Equal(t, "optional", homeData.UserExpensesCategories.OptionalExpenses.Name)
	require.Empty(t, homeData.UserExpensesCategories.MandatoryExpenses.Children)
	require.Empty(t, homeData.UserExpensesCategories.OptionalExpenses.Children)

	rows, err := container.Repo.LoadAllUserExpenses(container.Ctx, usr.ID)
	require.NoError(t, err)
	require.Len(t, rows, 2)
}
