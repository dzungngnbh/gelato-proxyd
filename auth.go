package proxyd

import (
	"context"
	"github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v5"
	"os"
	"strings"
)

// AuthStore is a struct that holds the keys for the authentication paths in url
// Example: http://endpoint?key=value
type AuthStore struct {
	keys map[string]string
}

func NewAuthStore() *AuthStore {
	keys := FetchKeysFromDb()
	if keys == nil {
		keys = make(map[string]string)
		log.Warn("No authentication keys found in database")
	}

	return &AuthStore{
		keys: keys,
	}
}

// Add adds a key value pair to the AuthStore
func (a *AuthStore) Add(key, value string) {
	if value == "" {
		return
	}

	a.keys[key] = value
}

// Remove removes a key from the AuthStore
func (a *AuthStore) Remove(key string) {
	delete(a.keys, key)
}

// Valid returns value if key is valid
func (a *AuthStore) Valid(key string, value string) string {
	if key == "" || value == "" {
		return ""
	}

	if a.keys[key] == value {
		return value
	}

	return ""
}

// Get returns value if key is valid
func (a *AuthStore) Get(key string) string {
	if key == "" {
		return ""
	}

	return a.keys[key]
}

// GetConn returns a connection to the database.
func GetConn() *pgx.Conn {
	pgUri := os.Getenv("DATABASE_URL")
	if pgUri == "" {
		log.Error("DATABASE_URL not set")
		return nil
	}

	conn, err := pgx.Connect(context.Background(), pgUri)
	if err != nil {
		log.Error("unable to connect to database", "err", err)
		return nil
	}

	return conn
}

// FetchKeysFromDb returns the keys from database.
func FetchKeysFromDb() map[string]string {
	conn := GetConn()
	if conn == nil {
		return nil
	}
	defer conn.Close(context.Background())

	// query auth keys from database
	var authKey string
	var authValue string
	var isDisabled bool

	rows, err := conn.Query(context.Background(), "select auth_key, auth_value, is_disabled from auth_keys")
	if err != nil {
		log.Error("query database failed", "err", err)
		return nil
	}

	res := make(map[string]string)
	_, err = pgx.ForEachRow(rows, []any{&authKey, &authValue, &isDisabled}, func() error {
		if !isDisabled {
			res[authKey] = authValue
		}
		return nil
	})

	return res
}

// UpsertNewAuthKey inserts a new auth key into the database if not existed otherwise re-enable it.
func UpsertNewAuthKey(authKey string) {
	conn := GetConn()
	defer conn.Close(context.Background())

	authKey = strings.TrimSpace(authKey)
	var isDisabled bool
	err := conn.QueryRow(context.Background(), "select auth_key, is_disabled from auth_keys where auth_key=$1", authKey).Scan(&authKey, &isDisabled)
	if err != nil {
		log.Error("query database failed", "err", err)
	} else {
		if isDisabled {
			_, err := conn.Exec(context.Background(), "update auth_keys set is_disabled=false where auth_key=$1", authKey)
			if err != nil {
				log.Error("update database failed", "err", err)
			}
		}
		return
	}

	_, err = conn.Exec(context.Background(), "insert into auth_keys (auth_key) values($1)", authKey)
	if err != nil {
		log.Error("insert database failed", "err", err)
		return
	}
}
