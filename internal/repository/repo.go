package repository

import "finance_tracker/internal/storage/database"

type Repo struct {
	db database.DBConnector
}

var AllTables = []string{
	// add table names here
}

func InitRepo(db database.DBConnector) *Repo {
	return &Repo{db: db}
}
