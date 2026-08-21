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

package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validBody = `{
  "struct": {
    "product": { "name": "user write new name here", "price": 0 },
    "user": { "name": "user write new name here", "email": "user write email here" }
  }
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
	got, err := Load(path, "")
	if err != nil {
		t.Fatalf("Load(%q) unexpected error: %v", path, err)
	}
	// Active structure is the first in sorted name order: product < user.
	if got.ActiveName != "product" {
		t.Errorf("Load(%q) ActiveName = %q ; want %q", path, got.ActiveName, "product")
	}
	if !strings.Contains(string(got.Body), `"price"`) {
		t.Errorf("Load(%q) Body = %q ; want the product structure", path, got.Body)
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	_, err := Load(path, "")
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "read body template") {
		t.Errorf("Load() error = %q ; want it to mention reading the template", err)
	}
}

func TestLoadBadJSON(t *testing.T) {
	path := writeTemplate(t, "{ this is not json ]")
	_, err := Load(path, "")
	if err == nil {
		t.Fatal("Load() expected error for bad JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parse body template") {
		t.Errorf("Load() error = %q ; want a parse message", err)
	}
}

func TestLoadMissingStructTag(t *testing.T) {
	path := writeTemplate(t, `{}`)
	_, err := Load(path, "")
	if err == nil {
		t.Fatal("Load({}) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing top-level struct tag") {
		t.Errorf("Load({}) error = %q ; want missing-struct-tag message", err)
	}
	if !strings.Contains(err.Error(), "not a body template") {
		t.Errorf("Load({}) error = %q ; want a not-a-body-template message", err)
	}
}

func TestLoadEmptyStructObject(t *testing.T) {
	path := writeTemplate(t, `{"struct": {}}`)
	_, err := Load(path, "")
	if err == nil {
		t.Fatal("Load({struct: {}}) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "struct tag is empty") {
		t.Errorf("Load({struct: {}}) error = %q ; want empty-struct message", err)
	}
}

func TestLoadWrongTypedStructTag(t *testing.T) {
	path := writeTemplate(t, `{"struct": ["product"]}`)
	_, err := Load(path, "")
	if err == nil {
		t.Fatal("Load(struct as array) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "must be a JSON object") {
		t.Errorf("Load(struct as array) error = %q ; want object-required message", err)
	}
}

// TestLoadForMethodBodyVisibility verifies only POST and PUT expose a body
// template; GET and DELETE get nil.
func TestLoadForMethodBodyVisibility(t *testing.T) {
	path := writeTemplate(t, validBody)

	for _, method := range []string{"post", "put"} {
		got, err := LoadForMethod(method, path, "")
		if err != nil {
			t.Fatalf("LoadForMethod(%q) unexpected error: %v", method, err)
		}
		if got == nil {
			t.Errorf("LoadForMethod(%q) returned nil body; want template", method)
		}
	}

	for _, method := range []string{"get", "delete"} {
		got, err := LoadForMethod(method, path, "")
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
	_, err := LoadForMethod("post", path, "")
	if err == nil {
		t.Fatal("LoadForMethod(post, missing) expected error, got nil")
	}
}

// TestLoadNamedStruct verifies -struct(name) selects the requested
// structure instead of the sorted default.
func TestLoadNamedStruct(t *testing.T) {
	path := writeTemplate(t, validBody)
	got, err := Load(path, "user")
	if err != nil {
		t.Fatalf("Load(user) unexpected error: %v", err)
	}
	if got.ActiveName != "user" {
		t.Errorf("Load(user) ActiveName = %q ; want %q", got.ActiveName, "user")
	}
	if !strings.Contains(string(got.Body), `"email"`) {
		t.Errorf("Load(user) Body = %q ; want the user structure", got.Body)
	}
}

// TestLoadUnknownStruct verifies a clear error when the selected structure
// does not exist in the template file.
func TestLoadUnknownStruct(t *testing.T) {
	path := writeTemplate(t, validBody)
	_, err := Load(path, "cart")
	if err == nil {
		t.Fatal("Load(cart) expected error, got nil")
	}
	if !strings.Contains(err.Error(), `struct "cart" does not exist`) {
		t.Errorf("Load(cart) error = %q ; want missing-struct message", err)
	}
}

// TestLoadForMethodPatchBody verifies PATCH carries a body like POST and PUT.
func TestLoadForMethodPatchBody(t *testing.T) {
	path := writeTemplate(t, validBody)
	got, err := LoadForMethod("patch", path, "")
	if err != nil {
		t.Fatalf("LoadForMethod(patch) unexpected error: %v", err)
	}
	if got == nil {
		t.Error("LoadForMethod(patch) returned nil body; want template")
	}
}

// TestLoadForMethodBodylessMethods verifies HEAD and OPTIONS carry no body.
func TestLoadForMethodBodylessMethods(t *testing.T) {
	path := writeTemplate(t, validBody)
	for _, method := range []string{"head", "options"} {
		got, err := LoadForMethod(method, path, "")
		if err != nil {
			t.Fatalf("LoadForMethod(%s) unexpected error: %v", method, err)
		}
		if got != nil {
			t.Errorf("LoadForMethod(%s) returned %q body; want nil", method, got)
		}
	}
}
