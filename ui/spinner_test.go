// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). It covers the spinner behaviour: it must cycle a
// frame and then clear its line, must write nothing while animation is
// disabled, and must tolerate a nil handle.
package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSpinnerDisabledWritesNothing(t *testing.T) {
	original := Enabled()
	SetEnabled(false)
	defer SetEnabled(original)

	var buf bytes.Buffer
	sp := NewSpinner(&buf)
	sp.Start()
	time.Sleep(5 * time.Millisecond)
	sp.Stop()

	if buf.Len() != 0 {
		t.Errorf("spinner with animation off wrote %q", buf.String())
	}
}

func TestSpinnerWritesFramesAndClears(t *testing.T) {
	original := Enabled()
	SetEnabled(true)
	defer SetEnabled(original)

	var buf bytes.Buffer
	sp := NewSpinner(&buf)
	sp.interval = 2 * time.Millisecond
	sp.Start()
	time.Sleep(20 * time.Millisecond)
	sp.Stop()

	out := buf.String()

	foundFrame := false
	for _, frame := range sp.frames {
		if strings.Contains(out, frame) {
			foundFrame = true
			break
		}
	}
	if !foundFrame {
		t.Errorf("spinner output %q contains no animation frame", out)
	}

	if !strings.HasSuffix(out, clearLine) {
		t.Errorf("spinner output %q does not end with a line clear", out)
	}
}

func TestNilSpinnerIsSafe(t *testing.T) {
	var sp *Spinner
	sp.Start()
	sp.Stop() // must not panic
}
