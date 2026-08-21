/*
Copyright 2026 Omar-BEDev

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"fmt"
	"reflect"
	"strings"
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
		{name: "leading spaces", input: "   new Req \"https://a.com\"", want: true},
		{name: "trailing spaces", input: "new req \"https://a.com\"   ", want: true},
		{name: "empty", input: "", want: false},
		{name: "whitespace only", input: "   ", want: false},
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
		name            string
		input           string
		wantURL         string
		wantMeths       []string
		wantErr         bool
		wantErrContains string
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
			name:      "mixed method forms",
			input:     `new req "https://a.com" -method=post -method(get,delete)`,
			wantURL:   "https://a.com",
			wantMeths: []string{"post", "get", "delete"},
		},
		{
			name:      "duplicate method flags",
			input:     `new req "https://a.com" -method(get) -method(get)`,
			wantURL:   "https://a.com",
			wantMeths: []string{"get"},
		},
		{
			name:      "trailing spaces",
			input:     `new req "https://a.com"   `,
			wantURL:   "https://a.com",
			wantMeths: []string{"post"},
		},
		{
			name:      "leading spaces",
			input:     `   new req "https://a.com"`,
			wantURL:   "https://a.com",
			wantMeths: []string{"post"},
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
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only input",
			input:   "     ",
			wantErr: true,
		},
		{
			name:            "unsupported scheme rejected",
			input:           `new req "ftp://files.example.com/x.zip"`,
			wantErr:         true,
			wantErrContains: `unsupported api link scheme "ftp"`,
		},
		{
			name:            "missing scheme rejected",
			input:           `new req "api.example.com/items"`,
			wantErr:         true,
			wantErrContains: "missing a scheme",
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
			input:   `new req "https://a.com" -method(trace)`,
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
				if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("ParseReq(%q) error = %q ; want it to contain %q", tt.input, err, tt.wantErrContains)
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

func TestParseReqHeadersAndStruct(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantHeaders     map[string]string
		wantStruct      string
		wantMethods     []string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "inline header flag",
			input:       `new req "https://a.com" -header("Authorization: Bearer tok")`,
			wantHeaders: map[string]string{"Authorization": "Bearer tok"},
		},
		{
			name:        "short header flag",
			input:       `new req "https://a.com" -H "X-Api-Key: abc"`,
			wantHeaders: map[string]string{"X-Api-Key": "abc"},
		},
		{
			name:  "multiple headers",
			input: `new req "https://a.com" -header("Accept: application/json") -H "X-Trace: 7"`,
			wantHeaders: map[string]string{
				"Accept":  "application/json",
				"X-Trace": "7",
			},
		},
		{
			name:        "canonical header name",
			input:       `new req "https://a.com" -header("content-type: text/plain")`,
			wantHeaders: map[string]string{"Content-Type": "text/plain"},
		},
		{
			name:            "inline header missing colon",
			input:           `new req "https://a.com" -header("just-a-name")`,
			wantErr:         true,
			wantErrContains: `invalid header`,
		},
		{
			name:            "short header without value",
			input:           `new req "https://a.com" -H`,
			wantErr:         true,
			wantErrContains: `-H requires`,
		},
		{
			name:            "header empty value",
			input:           `new req "https://a.com" -H "X-Empty:"`,
			wantErr:         true,
			wantErrContains: `invalid header`,
		},
		{
			name:       "struct selector",
			input:      `new req "https://a.com" -struct(user)`,
			wantStruct: "user",
		},
		{
			name:        "struct selector with methods",
			input:       `new req "https://a.com" -method(post,patch) -struct(product)`,
			wantStruct:  "product",
			wantMethods: []string{"post", "patch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReq(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseReq(%q) expected error, got %+v", tt.input, got)
				}
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("ParseReq(%q) error = %q ; want it to contain %q", tt.input, err, tt.wantErrContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseReq(%q) unexpected error: %v", tt.input, err)
			}
			if len(got.Headers) != len(tt.wantHeaders) {
				t.Fatalf("ParseReq(%q) headers = %v ; want %v", tt.input, got.Headers, tt.wantHeaders)
			}
			for name, value := range tt.wantHeaders {
				if got.Headers[name] != value {
					t.Errorf("ParseReq(%q) header %q = %q ; want %q", tt.input, name, got.Headers[name], value)
				}
			}
			if got.StructName != tt.wantStruct {
				t.Errorf("ParseReq(%q) StructName = %q ; want %q", tt.input, got.StructName, tt.wantStruct)
			}
			// Decision: when a case does not list methods the parser
			// defaults to POST, so nil expectations mean the default.
			want := tt.wantMethods
			if want == nil {
				want = []string{"post"}
			}
			if len(got.Methods) != len(want) {
				t.Fatalf("ParseReq(%q) methods = %v ; want %v", tt.input, got.Methods, want)
			}
			for i, m := range want {
				if got.Methods[i] != m {
					t.Errorf("ParseReq(%q) method %d = %q ; want %q", tt.input, i, got.Methods[i], m)
				}
			}
		})
	}
}

// TestParseReqExtendedMethods verifies the newly supported methods parse and
// normalise to lowercase.
func TestParseReqExtendedMethods(t *testing.T) {
	for _, m := range []string{"patch", "head", "options"} {
		got, err := ParseReq(fmt.Sprintf(`new req "https://a.com" -method(%s)`, m))
		if err != nil {
			t.Fatalf("ParseReq(method %s) unexpected error: %v", m, err)
		}
		if len(got.Methods) != 1 || got.Methods[0] != m {
			t.Errorf("ParseReq(method %s) methods = %v ; want [%s]", m, got.Methods, m)
		}
	}
}
