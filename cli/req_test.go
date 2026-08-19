// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). It contains unit tests for the new req command
// parser, following the go test convention required by the project
// contributing guidelines.
package cli

import (
	"reflect"
	"testing"
)

func TestIsReqCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "canonical", input: "new req \"https://a.com\"", want: true},
		{name: "capital new", input: "New req \"https://a.com\"", want: true},
		{name: "capital Req", input: "new Req \"https://a.com\"", want: true},
		{name: "no link", input: "new req", want: true},
		{name: "empty", input: "", want: false},
		{name: "only req", input: "req", want: false},
		{name: "other command", input: "login", want: false},
	}
	for _, tt := range tests {
		if got := IsReqCommand(tt.input); got != tt.want {
			t.Errorf("IsReqCommand(%q) = %v ; want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseReq(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantURL   string
		wantMeths []string
		wantErr   bool
	}{
		{
			name:      "defaults to post",
			input:     `new req "https://a.com"`,
			wantURL:   "https://a.com",
			wantMeths: []string{"post"},
		},
		{
			name:      "single method",
			input:     `new req "https://a.com" -method(get)`,
			wantURL:   "https://a.com",
			wantMeths: []string{"get"},
		},
		{
			name:      "many methods",
			input:     `new req "https://a.com" -method(post,get,put,delete)`,
			wantURL:   "https://a.com",
			wantMeths: []string{"post", "get", "put", "delete"},
		},
		{
			name:      "mixed case methods",
			input:     `new req "https://a.com" -method(POST,Get)`,
			wantURL:   "https://a.com",
			wantMeths: []string{"post", "get"},
		},
		{
			name:      "duplicates removed",
			input:     `new req "https://a.com" -method(get,post,get)`,
			wantURL:   "https://a.com",
			wantMeths: []string{"get", "post"},
		},
		{
			name:      "equal sign form",
			input:     `new req "https://a.com" -method=delete`,
			wantURL:   "https://a.com",
			wantMeths: []string{"delete"},
		},
		{
			name:      "unquoted link",
			input:     "new req https://a.com -method(put)",
			wantURL:   "https://a.com",
			wantMeths: []string{"put"},
		},
		{
			name:      "realistic query link",
			input:     `new req "https://api.example.com/v1/items?q=hello&page=2"`,
			wantURL:   "https://api.example.com/v1/items?q=hello&page=2",
			wantMeths: []string{"post"},
		},
		{
			name:    "link with whitespace rejected",
			input:   `new req "https://api.example.com/v1/items?q=a b"`,
			wantErr: true,
		},
		{
			name:    "missing link",
			input:   "new req",
			wantErr: true,
		},
		{
			name:    "empty link",
			input:   `new req ""`,
			wantErr: true,
		},
		{
			name:    "unsupported method",
			input:   `new req "https://a.com" -method(patch)`,
			wantErr: true,
		},
		{
			name:    "empty method list",
			input:   `new req "https://a.com" -method()`,
			wantErr: true,
		},
		{
			name:    "not a req command",
			input:   "login",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReq(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseReq(%q) expected error, got %+v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseReq(%q) unexpected error: %v", tt.input, err)
			}
			if got.URL != tt.wantURL {
				t.Errorf("ParseReq(%q) URL = %q ; want %q", tt.input, got.URL, tt.wantURL)
			}
			if !reflect.DeepEqual(got.Methods, tt.wantMeths) {
				t.Errorf("ParseReq(%q) Methods = %v ; want %v", tt.input, got.Methods, tt.wantMeths)
			}
		})
	}
}
