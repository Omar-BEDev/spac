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
	"bufio"
	"io"
	"os"

	"github.com/chzyer/readline"

	"spac/ui"
)

// lineReader is the single source of user input for the console loop. It is
// an interface so unit tests can feed scripted lines deterministically while
// the real binary uses a full terminal line editor.
type lineReader interface {
	// ReadLine returns the next trimmed-of-newline input line. It returns
	// io.EOF when the input stream is exhausted (Ctrl-D or a closed pipe).
	ReadLine() (string, error)
	// AddHistory records a submitted command so it can be recalled with the
	// Up/Down arrow keys. Implementations that record automatically (the
	// readline editor) may keep this as a no-op.
	AddHistory(line string)
	// Close releases terminal resources. It must be safe to call more than
	// once.
	Close()
}

// newLineReader picks the right input backend for the current environment:
// a full readline editor when stdin is an interactive character device, or a
// plain buffered scanner when stdin is piped or redirected.
func newLineReader() (lineReader, error) {
	if stdinIsTerminal() {
		return newInteractiveReader()
	}
	return newPlainReader(), nil
}

// stdinIsTerminal reports whether os.Stdin is an interactive character
// device rather than a pipe, a redirected file or /dev/null.
//
// Decision: ModeCharDevice is used instead of a third-party TTY probe so the
// fallback rule stays dependency-free and works the same on Linux, macOS and
// Termux.
func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// interactiveReader adapts a readline.Instance to the lineReader interface.
// It gives the user full line editing: Left/Right cursor movement, Backspace
// and Delete, Home/End, and Up/Down cycling through the session command
// history.
type interactiveReader struct {
	rl *readline.Instance
}

// newInteractiveReader builds the readline-backed editor. The prompt keeps
// the blue ">> " decoration, and the session history persists across runs in
// the user's config directory when one can be resolved.
//
// Decision: the persistent history file lives next to the action log under
// the user config directory; if that path cannot be determined the editor
// still starts fine with in-memory-only history, so interactive editing is
// never lost just because a directory is missing.
func newInteractiveReader() (*interactiveReader, error) {
	cfg := &readline.Config{
		Prompt: ui.Blue(">> "),
	}
	if file := readlineHistoryFile(); file != "" {
		cfg.HistoryFile = file
	}

	rl, err := readline.NewEx(cfg)
	if err != nil {
		return nil, err
	}
	return &interactiveReader{rl: rl}, nil
}

// ReadLine returns the next line from the terminal. A readline EOF (Ctrl-D)
// surfaces as io.EOF so the caller can exit uniformly across backends.
func (ir *interactiveReader) ReadLine() (string, error) {
	line, err := ir.rl.Readline()
	if err == readline.ErrInterrupt {
		// Decision: Ctrl-C cancels the current line instead of killing the
		// session; the user exits explicitly with exit, quit or Ctrl-D.
		return "", nil
	}
	return line, err
}

// AddHistory is a no-op: readline records every submitted line itself.
func (ir *interactiveReader) AddHistory(string) {}

// Close restores the terminal state. Double closes are tolerated because the
// underlying library ignores them.
func (ir *interactiveReader) Close() {
	_ = ir.rl.Close()
}

// plainReader is the non-TTY fallback built on bufio.Scanner. Piped input,
// redirected files and automated tests never hang waiting for a terminal.
type plainReader struct {
	scanner *bufio.Scanner
	history []string
}

// newPlainReader returns a scanner-backed reader over os.Stdin.
func newPlainReader() *plainReader {
	return &plainReader{scanner: bufio.NewScanner(os.Stdin)}
}

// ReadLine returns the next line or io.EOF once the stream is exhausted.
func (pr *plainReader) ReadLine() (string, error) {
	if !pr.scanner.Scan() {
		if err := pr.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return pr.scanner.Text(), nil
}

// AddHistory records the command in memory. There is no terminal to redraw,
// so recall keys are meaningless here; the slice exists so behaviour stays
// observable and testable in non-interactive sessions.
func (pr *plainReader) AddHistory(line string) {
	pr.history = append(pr.history, line)
}

// Close is a no-op for the scanner backend; stdin is not owned by spac.
func (pr *plainReader) Close() {}

// readlineHistoryFile resolves the persisted command-history location using
// the same user-directory rules as the action log. An empty result means no
// persistence this run.
func readlineHistoryFile() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil || home == "" {
			return ""
		}
		dir = home + "/.spac"
	} else {
		dir = dir + "/spac"
	}
	return dir + "/history.input"
}
