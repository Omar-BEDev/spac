// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). It covers the color helpers and the color/animation
// toggle, including the "no decorations when disabled" contract that keeps
// escape sequences out of tests and redirected output.
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
