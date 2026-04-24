package repository

import "finance_tracker/internal/storage/database"

type Repo struct {
	db database.DBConnector
}

var AllTables = []string{
	TableUsers,
	TableCurrencies,
}

func InitRepo(db database.DBConnector) *Repo {
	return &Repo{db: db}
}
