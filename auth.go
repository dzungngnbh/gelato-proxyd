package proxyd

import (
	"context"
	"fmt"
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

// UpsertNewAuthKey creates or updates an authentication key-value pair in the database.
// If the key exists, it will be re-enabled and its value updated.
// If the key doesn't exist, a new entry will be created.
//
// Parameters:
//   - authKey: The authentication key to create or update
//   - authValue: The value to associate with the auth key
//
// Returns:
//   - error: nil if successful, otherwise returns an error describing what went wrong
func UpsertNewAuthKey(authKey string, authValue string) error {
	// Validate input parameters
	if authKey == "" || authValue == "" {
		return fmt.Errorf("auth key and value cannot be empty")
	}

	// Clean input by removing leading/trailing whitespace
	authKey = strings.TrimSpace(authKey)

	// Get database connection
	conn := GetConn()
	if conn == nil {
		return fmt.Errorf("failed to establish database connection")
	}
	defer conn.Close(context.Background())

	// First, check if the key exists
	var existingKey string
	err := conn.QueryRow(
		context.Background(),
		"SELECT auth_key FROM auth_keys WHERE auth_key = $1",
		authKey,
	).Scan(&existingKey)

	// Handle existing key case
	if err == nil {
		// Key exists, update its value and enable it
		_, err := conn.Exec(
			context.Background(),
			`UPDATE auth_keys 
             SET is_disabled = false, 
                 auth_value = $1,
                 updated_at = CURRENT_TIMESTAMP
             WHERE auth_key = $2`,
			authValue,
			authKey,
		)
		if err != nil {
			return fmt.Errorf("failed to update existing auth key: %w", err)
		}
		return nil
	}

	// Handle non-existing key case
	if err != pgx.ErrNoRows {
		// Unexpected error during query
		return fmt.Errorf("failed to query auth key: %w", err)
	}

	// Key doesn't exist, create new entry
	_, err = conn.Exec(
		context.Background(),
		`INSERT INTO auth_keys 
         (auth_key, auth_value, is_disabled, created_at, updated_at)
         VALUES ($1, $2, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		authKey,
		authValue,
	)
	if err != nil {
		return fmt.Errorf("failed to insert new auth key: %w", err)
	}

	return nil
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
