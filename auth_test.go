package proxyd

import "testing"

func TestAuthStore(t *testing.T) {
	a := NewAuthStore()
	// test Set key
	a.Set("key", "value")

	// test Valid key returns value
	isValid := a.Valid("key", "value")
	if isValid == "" {
		t.Errorf("Expected true, got %v", isValid)
	}

	// remove key
	a.Delete("key")

	isValid = a.Valid("key", "value")
	if isValid != "" {
		t.Errorf("Expected empty string, got %v", isValid)
	}

	// test Get
	a.Set("key", "value")
	val := a.Get("key")
	if val != "value" {
		t.Errorf("Expected value, got %v", val)
	}
}
