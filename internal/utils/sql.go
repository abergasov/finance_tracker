package utils

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
)

// Querier is the common interface to execute queries on a DB, Tx, or Conn.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// GenerateUpdateSQL generates an UPDATE SQL statement with placeholders and parameters.
// - tableName: the name of the table to update.
// - fieldsValuesMapping: a map of column names to new values (the SET clause).
// - whereFieldsValuesMapping: a map of column names to values for the WHERE clause.
// Returns the SQL string and a slice of parameters.
func GenerateUpdateSQL(tableName string, fieldsValuesMapping, whereFieldsValuesMapping map[string]any) (sqlU string, params []any) {
	if len(fieldsValuesMapping) == 0 {
		return sqlU, params
	}
	setClauses := make([]string, 0, len(fieldsValuesMapping))
	params = make([]any, 0, len(fieldsValuesMapping)+len(whereFieldsValuesMapping))
	counter := 1

	// Build SET clauses and collect parameters
	for k, v := range fieldsValuesMapping {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, counter))
		params = append(params, v)
		counter++
	}

	// Build WHERE clauses and collect parameters
	whereClauses := make([]string, 0, len(whereFieldsValuesMapping))
	for k, v := range whereFieldsValuesMapping {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", k, counter))
		params = append(params, v)
		counter++
	}
	sqlU = fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setClauses, ", "))
	if len(whereClauses) > 0 {
		sqlU += fmt.Sprintf(" WHERE %s", strings.Join(whereClauses, " AND "))
	}

	return sqlU, params
}

func GenerateInsertSQL(tableName string, fieldsValuesMapping map[string]any) (sqlI string, params []any) {
	fields := make([]string, 0, len(fieldsValuesMapping))
	placeholders := make([]string, 0, len(fieldsValuesMapping))
	params = make([]any, 0, len(fieldsValuesMapping))
	counter := 1
	for k, v := range fieldsValuesMapping {
		params = append(params, v)
		fields = append(fields, k)
		placeholders = append(placeholders, fmt.Sprintf("$%d", counter))
		counter++
	}
	sqlI = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(fields, ", "), strings.Join(placeholders, ", "))
	return sqlI, params
}

// GenerateBulkInsertSQL This method generates a bulk insert SQL statement based on entity mapping
// will panic if entityList is empty
func GenerateBulkInsertSQL[T any](
	tableName string,
	entityList []T,
	entityProcessor func(entity T) map[string]any,
) (sqlI string, params []any) {
	// processor is based on map, so it random iteration. generate columns first
	columns := make([]string, 0, 10)
	for k := range entityProcessor(entityList[0]) {
		columns = append(columns, k)
	}

	// generate values
	counter := 1
	placeholders := make([]string, 0, len(entityList)*len(columns))
	params = make([]any, 0, len(entityList)*len(columns))
	for i := range entityList {
		sqlMapping := entityProcessor(entityList[i])
		localPlaceholders := make([]string, 0, len(columns))
		for j := range columns {
			params = append(params, sqlMapping[columns[j]])
			localPlaceholders = append(localPlaceholders, fmt.Sprintf("$%d", counter))
			counter++
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(localPlaceholders, ",")))
	}

	sqlI = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, strings.Join(columns, ","), strings.Join(placeholders, ","))
	return sqlI, params
}

func QueryRowsToStruct[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) ([]*T, error) {
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close() //nolint:errcheck
	res := make([]*T, 0, 100)
	rowScanner := sqlscan.NewRowScanner(rows)
	for rows.Next() {
		var t T
		if errS := rowScanner.Scan(&t); errS != nil {
			return nil, errS
		}
		res = append(res, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func QueryRowToStruct[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) (*T, error) {
	var t T
	if err := sqlscan.Get(ctx, conn, &t, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get row: %w", err)
	}
	return &t, nil
}

func QueryRowsPrimitive[T any](ctx context.Context, conn sqlscan.Querier, query string, params ...any) ([]T, error) {
	rows, err := conn.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // it's ok
	result := make([]T, 0, 1_000)
	for rows.Next() {
		var data T
		if errS := rows.Scan(&data); errS != nil {
			return nil, errS
		}
		result = append(result, data)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func QueryRowPrimitive[T any](ctx context.Context, conn Querier, query string, params ...any) (T, error) {
	var t T
	err := conn.QueryRowContext(ctx, query, params...).Scan(&t)
	return t, err
}
