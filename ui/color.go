// Package ui provides the small terminal helpers used by the spac console:
// ANSI color helpers and a minimal in-place "loading" spinner.
//
// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). Decisions are commented inline to make the
// reasoning behind them explicit.
package ui

import (
	"io"
	"os"
	"sync/atomic"
)

const (
	blue  = "\x1b[34m"
	red   = "\x1b[31m"
	reset = "\x1b[0m"
)

// colorEnabled mirrors whether terminal decorations (colors and animation)
// are active. It defaults to whether stdout looks like a real terminal and
// can be overridden with SetEnabled. Tests and piped runs call SetEnabled
// so escape sequences never leak into captured or redirected output.
var colorEnabled atomic.Bool

func init() {
	SetEnabled(isTerminal(os.Stdout))
}

// Blue wraps s in the blue ANSI color when decorations are enabled,
// otherwise returns s unchanged.
func Blue(s string) string {
	return wrap(s, blue)
}

// Red wraps s in the red ANSI color (used for error lines) when decorations
// are enabled, otherwise returns s unchanged.
func Red(s string) string {
	return wrap(s, red)
}

// SetEnabled toggles whether colors and the spinner animation are emitted.
func SetEnabled(enabled bool) {
	colorEnabled.Store(enabled)
}

// Enabled reports whether colors and the spinner animation are currently
// emitted.
func Enabled() bool {
	return colorEnabled.Load()
}

// wrap surrounds s with an ANSI color sequence unless decorations are off.
func wrap(s, color string) string {
	if !Enabled() {
		return s
	}
	return color + s + reset
}

// isTerminal reports whether w is an interactive character device (a TTY).
// Decision: when stdout is redirected (pipes, logs, CI) we turn decorations
// off so the console stays portable and grep/piping stays clean.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
