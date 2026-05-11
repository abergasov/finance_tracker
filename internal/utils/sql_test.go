package utils_test

import (
	"finance_tracker/internal/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

type Fruit struct {
	name   string
	amount int
}

func TestGenerateUpdateSQL(t *testing.T) {
	tests := []struct {
		name           string
		expectedSQL    []string
		expectedParams [][]any

		tableName                string
		fieldsValuesMapping      map[string]any
		whereFieldsValuesMapping map[string]any
	}{
		{
			name:      "Simple update with where clause",
			tableName: "users",
			fieldsValuesMapping: map[string]any{
				"name": "Alice",
				"age":  30,
			},
			whereFieldsValuesMapping: map[string]any{
				"id": 123,
			},
			expectedSQL: []string{
				"UPDATE users SET name = $1, age = $2 WHERE id = $3",
				"UPDATE users SET age = $1, name = $2 WHERE id = $3",
			},
			expectedParams: [][]any{
				{"Alice", 30, 123},
				{30, "Alice", 123},
			},
		},
		{
			name:      "Update without where clause",
			tableName: "products",
			fieldsValuesMapping: map[string]any{
				"price": 19.99,
			},
			whereFieldsValuesMapping: map[string]any{},
			expectedSQL:              []string{"UPDATE products SET price = $1"},
			expectedParams: [][]any{
				{
					19.99,
				},
			},
		},
		{
			name:      "Update with multiple where conditions",
			tableName: "orders",
			fieldsValuesMapping: map[string]any{
				"status": "shipped",
			},
			whereFieldsValuesMapping: map[string]any{
				"user_id":  456,
				"order_id": 789,
			},
			expectedSQL: []string{
				"UPDATE orders SET status = $1 WHERE user_id = $2 AND order_id = $3",
				"UPDATE orders SET status = $1 WHERE order_id = $2 AND user_id = $3",
			},
			expectedParams: [][]any{
				{"shipped", 456, 789},
				{"shipped", 789, 456},
			},
		},
		{
			name:                "Empty fieldsValuesMapping",
			tableName:           "employees",
			fieldsValuesMapping: map[string]any{},
			whereFieldsValuesMapping: map[string]any{
				"employee_id": 101,
			},
			expectedSQL: []string{""},
			expectedParams: [][]any{
				nil,
			},
		},
		{
			name:      "Special characters in identifiers",
			tableName: `"user-accounts"`,
			fieldsValuesMapping: map[string]any{
				`"first-name"`: "Bob",
				`"last-name"`:  "Smith",
			},
			whereFieldsValuesMapping: map[string]any{
				`"user-id"`: 202,
			},
			expectedSQL: []string{
				`UPDATE "user-accounts" SET "first-name" = $1, "last-name" = $2 WHERE "user-id" = $3`,
				`UPDATE "user-accounts" SET "last-name" = $1, "first-name" = $2 WHERE "user-id" = $3`,
			},
			expectedParams: [][]any{
				{"Bob", "Smith", 202},
				{"Smith", "Bob", 202},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params := utils.GenerateUpdateSQL(tt.tableName, tt.fieldsValuesMapping, tt.whereFieldsValuesMapping)
			found := false
			for i := range tt.expectedSQL {
				if query == tt.expectedSQL[i] {
					require.Equal(t, tt.expectedParams[i], params)
					found = true
				}
			}
			require.True(t, found, "unexpected query: %s", query)
		})
	}
}

func TestGenerateInsertSQL(t *testing.T) {
	t.Parallel()

	result, params := utils.GenerateInsertSQL("fruits", map[string]any{
		"name": "amount",
	})
	require.Equal(t, "INSERT INTO fruits (name) VALUES ($1)", result)
	require.Len(t, params, 1)
	require.Equal(t, "amount", params[0])

	result, params = utils.GenerateInsertSQL("fruits", map[string]any{
		"name":  "amount",
		"count": 1,
	})
	require.Len(t, params, 2)
	switch result {
	case "INSERT INTO fruits (name, count) VALUES ($1, $2)":
		require.Equal(t, "amount", params[0])
		require.Equal(t, 1, params[1])
	case "INSERT INTO fruits (count, name) VALUES ($1, $2)":
		require.Equal(t, "amount", params[1])
		require.Equal(t, 1, params[0])
	default:
		t.Errorf("unexpected result: %s", result)
	}
}

func TestGenerateBulkInsertSQL(t *testing.T) {
	res, params := utils.GenerateBulkInsertSQL[Fruit]("sample", []Fruit{
		{"Apple", 10},
		{"Pear", 100},
		{"Cherry", 36},
		{"Banana", 4},
		{"Apricot", 99},
	}, func(entity Fruit) map[string]any {
		return map[string]any{
			"name":   entity.name,
			"amount": entity.amount,
		}
	})
	t.Log(res)
	valid := res == "INSERT INTO sample (name,amount) VALUES ($1,$2),($3,$4),($5,$6),($7,$8),($9,$10)" ||
		res == "INSERT INTO sample (amount,name) VALUES ($1,$2),($3,$4),($5,$6),($7,$8),($9,$10)"
	require.True(t, valid)
	require.Len(t, params, 10)
}
