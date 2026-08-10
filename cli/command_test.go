package cli

import (
	"testing"
)

func TestIsAvailableCommand(t *testing.T) {
	tests := []struct {
		name     string
		excepted bool
	}{
		{
			name:     "post",
			excepted: true,
		},
		{
			name:     "get",
			excepted: true,
		},
		{
			name: "put", 
			excepted: true,
		},
		{
			name: "delete", 
			excepted: true,
		},
		{
			name: "new Req", 
			excepted: true,
		},
		{
			name: "", 
			excepted: false,
		},
		{
			name: "POST", 
			excepted: false,
		},
		{
			name: " post", 
			excepted: false,
		},
		{
			name: "post ", 
			excepted: false,
		},
		{
			name: "new req", 
			excepted: false,
		},
		{
			name: "new Req ", 
			excepted: false,
		},
		{
			name: "gert", 
			excepted: false,
		},
		{
			name: "api", 
			excepted: false,
		},
		{
			name: "method", 
			excepted: false,
		},
		{
			name: "cmd", 
			excepted: false,
		},
	}
	for _, tt := range tests {
		result := IsAvailableCommand(tt.name)
		if result != tt.excepted {
			t.Errorf("IsAvailableCommand(%q) = %v ; excepted (%v)", tt.name, result, tt.excepted)
		}
	}
}
