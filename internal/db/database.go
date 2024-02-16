package db

import "database/sql"

func Connect() (*sql.DB, error) {
	driverName := "sqlite3"
	dataSourceName := "data/sqlite/main.db"
	return sql.Open(driverName, dataSourceName)
}
