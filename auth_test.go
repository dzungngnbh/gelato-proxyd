package proxyd

import "testing"

func TestAuthStore(t *testing.T) {
	a := NewAuthStore()
	a.Add("key", "value")
	isValid := a.Valid("key", "value")
	if !isValid {
		t.Errorf("Expected true, got %v", isValid)
	}

	// remove key
	a.Remove("key")

	isValid = a.Valid("key", "value")
	if isValid {
		t.Errorf("Expected false, got %v", isValid)
	}
}
