package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Connect() (*sql.DB, error) {
	driverName := "sqlite3"
	dataSourceName := "data/sqlite/main.db"
	return sql.Open(driverName, dataSourceName)
}
