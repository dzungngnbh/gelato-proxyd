package proxyd

// AuthStore is a struct that holds the keys for the authentication paths in url
// Example: http://endpoint?key=value
type AuthStore struct {
	keys map[string]string
}

func NewAuthStore() *AuthStore {
	return &AuthStore{
		keys: make(map[string]string),
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

// Valid checks if the key value pair is valid
func (a *AuthStore) Valid(key string, value string) bool {
	if key == "" || value == "" {
		return false
	}

	return a.keys[key] == value
}
