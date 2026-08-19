// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). It covers the run -tests command parser.
package cli

import "testing"

func TestIsRunTestsCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "with path", input: `run -tests "../tests.json"`, want: true},
		{name: "no path", input: "run -tests", want: true},
		{name: "capital", input: "Run -Tests \"a.json\"", want: true},
		{name: "empty", input: "", want: false},
		{name: "typo", input: "run tests x.json", want: false},
		{name: "other command", input: "new req \"https://a.com\"", want: false},
	}
	for _, tt := range tests {
		if got := IsRunTestsCommand(tt.input); got != tt.want {
			t.Errorf("IsRunTestsCommand(%q) = %v ; want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseRunTests(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "quoted path", input: `run -tests "../tests.json"`, want: "../tests.json"},
		{name: "plain path", input: "run -tests tests.json", want: "tests.json"},
		{name: "path with spaces", input: `run -tests "../my tests/case.json"`, want: "../my tests/case.json"},
		{name: "missing path", input: "run -tests", wantErr: true},
		{name: "empty input", input: "", wantErr: true},
		{name: "not a run command", input: "login", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRunTests(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseRunTests(%q) expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRunTests(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseRunTests(%q) = %q ; want %q", tt.input, got, tt.want)
			}
		})
	}
}
