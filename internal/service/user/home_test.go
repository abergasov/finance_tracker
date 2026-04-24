package user_test

import (
	"database/sql"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/service/user"
	"testing"

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
					ChildExpense: &entities.UserExpensesCategory{
						ID:   2,
						Name: "rent",
					},
				},
				OptionalExpenses: entities.UserExpensesCategory{
					ID:   3,
					Name: "optional",
					ChildExpense: &entities.UserExpensesCategory{
						ID:   4,
						Name: "travel",
						ChildExpense: &entities.UserExpensesCategory{
							ID:   5,
							Name: "hotel",
						},
					},
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
		"missing parent": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "mandatory"},
				{ID: 2, Name: "rent", ParentID: sql.NullInt64{Int64: 404, Valid: true}},
			},
			wantErr: "missing parent category: id=2 parent_id=404",
		},
		"parent has more than one child": {
			rows: []*entities.UserExpensesCategoryDB{
				{ID: 1, Name: "mandatory"},
				{ID: 2, Name: "rent", ParentID: sql.NullInt64{Int64: 1, Valid: true}},
				{ID: 3, Name: "taxes", ParentID: sql.NullInt64{Int64: 1, Valid: true}},
			},
			wantErr: "parent category has more than one child: parent_id=1",
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
