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

// Package history provides a simple way to record user actions in a
// local log file so the tool keeps a trace of what happened.
package history

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	historyFileName = "history.log"
)

// logFilePath is the path of the history log file. It is a package
// variable instead of a constant so tests can point it to a temporary
// location without touching the real log.
//
// Decision: the log lives in the user's own config directory (with a home
// directory fallback) instead of next to the executable. When the binary is
// installed in a system path such as /usr/local/bin the executable directory
// is not writable by normal users, so the previous approach broke logging.
var defaultLogFilePath = func() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "spac", historyFileName)
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".spac", historyFileName)
	}
	// Last resort: the current working directory.
	return historyFileName
}()

var logFilePath = defaultLogFilePath

// SetLogFilePath overrides the history log file location so tests in other
// packages (for example main) can point the log at a temporary file. Passing
// an empty path restores the default location derived from the user's
// config (or home) directory.
//
// Decision: exported on purpose as a test seam; the real location logic stays
// unexported in defaultLogFilePath.
func SetLogFilePath(path string) {
	if path == "" {
		logFilePath = defaultLogFilePath
		return
	}
	logFilePath = path
}

// LogEntry represents a single history log entry.
type LogEntry struct {
	Date   string
	Action string
}

// LogAction appends an action with its current date to the history log.
// The stored line format is "YYYY/MM/DD action".
//
// It returns an error if the log file cannot be opened or written to.
func LogAction(action string) error {
	// Decision: the parent directory (for example ~/.config/spac) may not
	// exist on a fresh install, so it is created on demand before opening
	// the log file.
	if dir := filepath.Dir(logFilePath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create history log directory: %w", err)
		}
	}

	currentDate := time.Now().Format("2006/01/02")
	entry := fmt.Sprintf("%s %s\n", currentDate, action)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open history log file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("write to history log file: %w", err)
	}

	return nil
}

// LogNewReq records a "new req" action in the history log.
func LogNewReq() error {
	return LogAction("new req")
}

// LogLogin records a "login" action in the history log.
func LogLogin() error {
	return LogAction("login")
}
