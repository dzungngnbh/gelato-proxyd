package proxyd

import "testing"

func TestAuthStore(t *testing.T) {
	a := NewAuthStore()
	// test Add key
	a.Add("key", "value")

	// test Valid key returns value
	isValid := a.Valid("key", "value")
	if isValid == "" {
		t.Errorf("Expected true, got %v", isValid)
	}

	// remove key
	a.Remove("key")

	isValid = a.Valid("key", "value")
	if isValid != "" {
		t.Errorf("Expected empty string, got %v", isValid)
	}

	// test Get
	a.Add("key", "value")
	val := a.Get("key")
	if val != "value" {
		t.Errorf("Expected value, got %v", val)
	}
}
