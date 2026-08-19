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
