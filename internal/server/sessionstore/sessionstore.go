package sessionstore

import (
	"os"

	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

func GetSessionStore(secure bool) (*sqlitestore.SqliteStore, error) {
	secretKey := os.Getenv("SECRET_KEY")

	endpoint := "./data/sqlite/sessions.db"
	tableName := "sessions"
	path := "/"
	maxAge := 3600
	keyPairs := []byte(secretKey)
	store, err := sqlitestore.NewSqliteStore(endpoint, tableName, path, maxAge, keyPairs)
	if err != nil {
		return nil, err
	}
	sessionOptions := &sessions.Options{
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
	}
	sessionOptions.Secure = secure

	store.Options = sessionOptions
	return store, nil
}
