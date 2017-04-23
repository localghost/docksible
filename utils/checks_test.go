package utils

import "testing"

func TestInStringSlice(t *testing.T) {
	slice := []string{"foo", "bar", "bazinga"}
	expected := "bar"

	if !InStringSlice(expected, slice) {
		t.Fatalf("String '%s' not found in slice '%s'.\n", expected, slice)
	}
}

func TestNotInStringSlice(t *testing.T) {
	slice := []string{"foo", "bazinga"}
	notExpected := "bar"

	if InStringSlice(notExpected, slice) {
		t.Fatalf("Unexpected string '%s' found in slice '%s'.\n", notExpected, slice)
	}
}
