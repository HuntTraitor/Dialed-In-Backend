package assert

import (
	"strings"
	"testing"
)

// Equal checks if two values are equal to each other (expected, actual)
func Equal[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if actual != expected {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func NilError[T comparable](t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("expected nil, got %v", actual)
	}
}

func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("expected %s to contain %s", actual, expectedSubstring)
	}

}
