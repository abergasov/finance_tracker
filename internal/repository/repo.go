package repository

import (
	"errors"
	"finance_tracker/internal/storage/database"
	"strings"

	"github.com/lib/pq"
)

type Repo struct {
	db database.DBConnector
}

var AllTables = []string{
	TableCategoryExpenses,
	TableUsers,
	TableCurrencies,
}

func InitRepo(db database.DBConnector) *Repo {
	return &Repo{db: db}
}

// isUniqueViolation reports whether err is a PostgreSQL unique-constraint
// violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "23505") {
		return true
	}
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
