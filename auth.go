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

// Set adds a key value pair to the AuthStore
func (a *AuthStore) Set(key, value string) {
	if value == "" {
		return
	}

	a.keys[key] = value
}

// Delete removes a key from the AuthStore
func (a *AuthStore) Delete(key string) {
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
func UpsertNewAuthKey(authKey string, authValue string) {
	if authKey == "" || authValue == "" {
		return
	}

	conn := GetConn()
	defer conn.Close(context.Background())

	authKey = strings.TrimSpace(authKey)
	err := conn.QueryRow(context.Background(), "select auth_key  from auth_keys where auth_key=$1", authKey).Scan(&authKey)
	if err == nil {
		// re-enable the key if it is disabled and set new value
		_, err := conn.Exec(context.Background(), "update auth_keys set is_disabled=false, auth_value=$1 where auth_key=$2", authValue, authKey)
		if err != nil {
			log.Error("update database failed", "err", err)
		}
		return
	}

	_, err = conn.Exec(context.Background(), "insert into auth_keys (auth_key, auth_value) values($1, $2)", authKey, authValue)
	if err != nil {
		log.Error("insert database failed", "err", err)
		return
	}
}

// DisableAuthKey disables an auth key in the database.
func DisableAuthKey(authKey string) {
	if authKey == "" {
		return
	}

	conn := GetConn()
	defer conn.Close(context.Background())

	authKey = strings.TrimSpace(authKey)
	_, err := conn.Exec(context.Background(), "update auth_keys set is_disabled=true where auth_key=$1", authKey)
	if err != nil {
		log.Error("update database failed", "err", err)
	}
}
