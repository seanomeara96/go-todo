package sessionstore

import (
	"fmt"
	"os"

	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

func GetSessionStore() (*sqlitestore.SqliteStore, error) {
	secretKey := os.Getenv("SECRET_KEY")
	env := os.Getenv("ENV")
	if secretKey == "" || env == "" {
		return nil, fmt.Errorf("missing env vars")
	}

	endpoint := "./sessions.db"
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
	if env == "prod" {
		sessionOptions.Secure = true
	}
	store.Options = sessionOptions
	return store, nil
}
