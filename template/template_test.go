// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). It covers body template loading: missing file,
// bad JSON, and the POST/PUT body-visibility rule.
package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validBody = `{
  "name": "user write new name here",
  "description": "user write description"
}`

func writeTemplate(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	return path
}

func TestLoadValid(t *testing.T) {
	path := writeTemplate(t, validBody)
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q) unexpected error: %v", path, err)
	}
	if !strings.Contains(string(got), "name") {
		t.Errorf("Load(%q) = %q ; want template content", path, got)
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "read body template") {
		t.Errorf("Load() error = %q ; want it to mention reading the template", err)
	}
}

func TestLoadBadJSON(t *testing.T) {
	path := writeTemplate(t, "{ this is not json ]")
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for bad JSON, got nil")
	}
	if !strings.Contains(err.Error(), "not valid JSON") {
		t.Errorf("Load() error = %q ; want a JSON parse message", err)
	}
}

// TestLoadForMethodBodyVisibility verifies only POST and PUT expose a body
// template; GET and DELETE get nil.
func TestLoadForMethodBodyVisibility(t *testing.T) {
	path := writeTemplate(t, validBody)

	for _, method := range []string{"post", "put"} {
		got, err := LoadForMethod(method, path)
		if err != nil {
			t.Fatalf("LoadForMethod(%q) unexpected error: %v", method, err)
		}
		if got == nil {
			t.Errorf("LoadForMethod(%q) returned nil body; want template", method)
		}
	}

	for _, method := range []string{"get", "delete"} {
		got, err := LoadForMethod(method, path)
		if err != nil {
			t.Fatalf("LoadForMethod(%q) unexpected error: %v", method, err)
		}
		if got != nil {
			t.Errorf("LoadForMethod(%q) returned %q body; want nil", method, got)
		}
	}
}

func TestLoadForMethodPostMissingTemplate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	_, err := LoadForMethod("post", path)
	if err == nil {
		t.Fatal("LoadForMethod(post, missing) expected error, got nil")
	}
}
