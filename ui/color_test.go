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

package ui

import (
	"strings"
	"testing"
)

func TestBlueEnabled(t *testing.T) {
	original := Enabled()
	SetEnabled(true)
	defer SetEnabled(original)

	got := Blue("ok")
	if !strings.HasPrefix(got, blue) || !strings.HasSuffix(got, reset) {
		t.Errorf("Blue(%q) = %q ; want %q...%q wrapper", "ok", got, blue, reset)
	}
	if !strings.Contains(got, "ok") {
		t.Errorf("Blue(%q) lost the payload: %q", "ok", got)
	}
}

func TestRedEnabled(t *testing.T) {
	original := Enabled()
	SetEnabled(true)
	defer SetEnabled(original)

	got := Red("bad")
	if !strings.HasPrefix(got, red) || !strings.HasSuffix(got, reset) {
		t.Errorf("Red(%q) = %q ; want %q...%q wrapper", "bad", got, red, reset)
	}
	if !strings.Contains(got, "bad") {
		t.Errorf("Red(%q) lost the payload: %q", "bad", got)
	}
}

func TestColorsDisabledPassThrough(t *testing.T) {
	original := Enabled()
	SetEnabled(false)
	defer SetEnabled(original)

	if got := Blue("ok"); got != "ok" {
		t.Errorf("Blue() with decorations off = %q ; want %q", got, "ok")
	}
	if got := Red("bad"); got != "bad" {
		t.Errorf("Red() with decorations off = %q ; want %q", got, "bad")
	}
	if strings.Contains(Blue("ok"), "\x1b") {
		t.Error("Blue() leaked an escape sequence while decorations are off")
	}
}
