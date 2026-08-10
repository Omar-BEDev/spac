package cli

import (
	"testing"
)

func TestIsAvailableCommand(t *testing.T) {
	tests := []struct {
		name     string
		excepted bool
	}{
		{name: "post", excepted: true},
		{name: "", excepted: false},
	}
	for _, tt := range tests {
		result := IsAvailableCommand(tt.name)
		if result != tt.excepted {
			t.Errorf("IsAvailableCommand(%q) = %v ; excepted (%v)", tt.name, result, tt.excepted)
		}
	}
}
