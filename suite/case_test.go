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

package suite

import (
	"strings"
	"testing"
)

func TestParseValid(t *testing.T) {
	data := []byte(`{
		"tests": [
			{"method": "get", "url": "https://api.example.com/health"},
			{"method": "post", "url": "https://api.example.com/users",
			 "body": {"name": "a", "price": 0}}
		]
	}`)
	cases, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("Parse() returned %d cases ; want 2", len(cases))
	}
	if cases[0].Method != "get" || cases[0].URL != "https://api.example.com/health" {
		t.Errorf("Parse() case 0 = %+v ; want get/health", cases[0])
	}
	if string(cases[1].Body) == "" {
		t.Error("Parse() case 1 Body empty ; want raw JSON kept")
	}
}

func TestParseBadJSON(t *testing.T) {
	_, err := Parse([]byte("not json at all"))
	if err == nil {
		t.Fatal("Parse() expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parse tests file") {
		t.Errorf("Parse() error = %q ; want it to mention parsing", err)
	}
}

func TestParseMissingTestsTag(t *testing.T) {
	data := []byte(`{}`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("Parse({}) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing top-level tests tag") {
		t.Errorf("Parse({}) error = %q ; want missing-tag message", err)
	}
}

func TestParseWrongTypedTestsTag(t *testing.T) {
	data := []byte(`{"tests": {"method": "get"}}`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("Parse(wrong-typed tests) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "tests tag must be a list") {
		t.Errorf("Parse(wrong-typed tests) error = %q ; want list-required message", err)
	}
}

func TestParseEmptyCases(t *testing.T) {
	_, err := Parse([]byte(`{"tests": []}`))
	if err == nil {
		t.Fatal("Parse() expected error for empty case list, got nil")
	}
}

func TestParseUnsupportedMethod(t *testing.T) {
	_, err := Parse([]byte(`{"tests": [{"method": "patch", "url": "https://a.com"}]}`))
	if err == nil {
		t.Fatal("Parse() expected error for unsupported method, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported method") {
		t.Errorf("Parse() error = %q ; want an unsupported-method message", err)
	}
}

func TestParseMissingURL(t *testing.T) {
	_, err := Parse([]byte(`{"tests": [{"method": "get", "url": "  "}]}`))
	if err == nil {
		t.Fatal("Parse() expected error for missing url, got nil")
	}
	if !strings.Contains(err.Error(), "missing url") {
		t.Errorf("Parse() error = %q ; want a missing-url message", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/dir/tests.json")
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "read tests file") {
		t.Errorf("Load() error = %q ; want it to mention reading the file", err)
	}
}
