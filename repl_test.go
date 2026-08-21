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

package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"spac/history"
)

// mockReader is a scripted lineReader: it replays canned lines and records
// everything runConsole submits to history, so REPL behaviour can be tested
// without a terminal.
type mockReader struct {
	lines   []string
	pos     int
	history []string
}

func (m *mockReader) ReadLine() (string, error) {
	if m.pos >= len(m.lines) {
		return "", io.EOF
	}
	line := m.lines[m.pos]
	m.pos++
	return line, nil
}

func (m *mockReader) AddHistory(line string) {
	m.history = append(m.history, line)
}

func (m *mockReader) Close() {}

// TestRunConsoleExecutesCommandsInOrder verifies multiple commands piped into
// the console all execute until EOF ends the session.
func TestRunConsoleExecutesCommandsInOrder(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "history.log")
	history.SetLogFilePath(logPath)

	server := okServer(t)
	mock := &mockReader{lines: []string{
		`new req "` + server.URL + `" -method(get)`,
		`unknown-cmd-xyz`,
	}}

	out := captureStdout(t, func() { runConsole(mock) })
	if !strings.Contains(out, "-> 200 OK") {
		t.Errorf("expected executed request in output, got %q", out)
	}
	if !strings.Contains(out, "unknown command") {
		t.Errorf("expected unknown-command message, got %q", out)
	}
	if !strings.Contains(out, "bye") {
		t.Errorf("EOF should end the session with goodbye, got %q", out)
	}
}

// TestRunConsoleExitAndQuit verifies both spellings end the session with the
// goodbye line and that later lines are never read.
func TestRunConsoleExitAndQuit(t *testing.T) {
	for _, cmd := range []string{"exit", "quit", "EXIT", "Quit"} {
		mock := &mockReader{lines: []string{cmd, "never-read"}}
		out := captureStdout(t, func() { runConsole(mock) })
		if !strings.Contains(out, "bye") {
			t.Errorf("runConsole(%q) expected goodbye, got %q", cmd, out)
		}
		if mock.pos != 1 {
			t.Errorf("runConsole(%q) read %d lines ; want exactly 1", cmd, mock.pos)
		}
	}
}

// TestRunConsoleBlankLinesAreIgnored verifies empty and whitespace-only lines
// consume no command processing and do not end the session.
func TestRunConsoleBlankLinesAreIgnored(t *testing.T) {
	mock := &mockReader{lines: []string{"", "   ", "\t", "exit"}}

	out := captureStdout(t, func() { runConsole(mock) })
	if strings.Contains(out, "unknown command") {
		t.Errorf("blank lines must not be dispatched, got %q", out)
	}
	if !strings.Contains(out, "bye") {
		t.Errorf("session should continue past blanks to exit, got %q", out)
	}
}

// TestRunConsoleRecordsSessionHistory verifies every submitted non-blank
// command is recorded for arrow-key recall.
func TestRunConsoleRecordsSessionHistory(t *testing.T) {
	server := okServer(t)
	mock := &mockReader{lines: []string{
		"",
		`new req "` + server.URL + `" -method(get)`,
		"run -tests",
		"exit",
	}}

	captureStdout(t, func() { runConsole(mock) })

	want := []string{
		`new req "` + server.URL + `" -method(get)`,
		"run -tests",
	}
	if len(mock.history) != len(want) {
		t.Fatalf("history = %v ; want %v", mock.history, want)
	}
	for i, line := range want {
		if mock.history[i] != line {
			t.Errorf("history[%d] = %q ; want %q", i, mock.history[i], line)
		}
	}
}

// TestRunConsoleReadError verifies an input error other than EOF stops the
// loop instead of spinning forever.
func TestRunConsoleReadError(t *testing.T) {
	reader := &errorReader{err: errors.New("simulated input failure")}

	out := captureStdout(t, func() { runConsole(reader) })
	if !strings.Contains(out, "input:") || !strings.Contains(out, "simulated input failure") {
		t.Errorf("expected input error surfaced, got %q", out)
	}
}

// errorReader always fails; it proves ReadLine errors terminate the loop.
type errorReader struct{ err error }

func (e *errorReader) ReadLine() (string, error) { return "", e.err }
func (e *errorReader) AddHistory(string)         {}
func (e *errorReader) Close()                    {}

// TestStdinIsTerminalWithPipedInput verifies the TTY probe reports false when
// stdin is redirected from a file, which is how CI and tests consume spac.
func TestStdinIsTerminalWithPipedInput(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "stdin-*")
	if err != nil {
		t.Fatalf("CreateTemp() error: %v", err)
	}
	defer file.Close()

	original := os.Stdin
	os.Stdin = file
	defer func() { os.Stdin = original }()

	if stdinIsTerminal() {
		t.Error("stdinIsTerminal() = true for a regular file ; want false")
	}

	reader, err := newLineReader()
	if err != nil {
		t.Fatalf("newLineReader() unexpected error: %v", err)
	}
	if _, ok := reader.(*plainReader); !ok {
		t.Errorf("newLineReader() returned %T ; want *plainReader for non-TTY stdin", reader)
	}
	reader.Close()
}

// TestPlainReaderReplay verifies the fallback backend replays every line and
// then reports io.EOF so piped sessions terminate cleanly.
func TestPlainReaderReplay(t *testing.T) {
	input := "first\n\n  second  \n"
	file, err := os.CreateTemp(t.TempDir(), "stdin-*")
	if err != nil {
		t.Fatalf("CreateTemp() error: %v", err)
	}
	defer file.Close()
	if _, err := file.WriteString(input); err != nil {
		t.Fatalf("WriteString() error: %v", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Seek() error: %v", err)
	}

	original := os.Stdin
	os.Stdin = file
	defer func() { os.Stdin = original }()

	reader := newPlainReader()
	want := []string{"first", "", "  second  "}
	for i, line := range want {
		got, err := reader.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine(%d) unexpected error: %v", i, err)
		}
		if got != line {
			t.Errorf("ReadLine(%d) = %q ; want %q", i, got, line)
		}
	}
	if _, err := reader.ReadLine(); err != io.EOF {
		t.Errorf("final ReadLine() error = %v ; want io.EOF", err)
	}
}
