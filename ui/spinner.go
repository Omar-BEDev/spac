package ui

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). Decisions are commented inline to make the
// reasoning behind them explicit.

// clearLine is printed to erase the current terminal line and return the
// cursor to its start. It is a widely supported ANSI sequence.
const clearLine = "\r\x1b[2K"

// Spinner draws a tiny "|/-\" animation in place while a blocking call, such
// as network.Send, runs. It is minimal and terminal-safe: frames are drawn
// with a carriage return instead of spawning a new terminal line.
type Spinner struct {
	writer   io.Writer
	frames   []string
	interval time.Duration

	stop chan struct{}
	done chan struct{}
	mu   sync.Mutex
}

// NewSpinner returns a stopped spinner that will write to w once started.
func NewSpinner(w io.Writer) *Spinner {
	return &Spinner{
		writer:   w,
		frames:   []string{"|", "/", "-", "\\"},
		interval: 75 * time.Millisecond,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Start begins drawing the spinner. When decorations are disabled it is a
// no-op, so redirected output and tests never see escape sequences.
func (s *Spinner) Start() {
	if s == nil || !Enabled() {
		return
	}
	go s.run()
}

// run ticks the spinner frames until Stop is called.
func (s *Spinner) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for frame := 0; ; frame++ {
		select {
		case <-s.stop:
			close(s.done)
			return
		case <-ticker.C:
			s.write(clearLine + s.frames[frame%len(s.frames)])
		}
	}
}

// Stop halts the spinner and erases its line. It is a no-op for a spinner
// that was never started or when decorations are disabled.
func (s *Spinner) Stop() {
	if s == nil || !Enabled() {
		return
	}
	close(s.stop)
	<-s.done
	// The ticker goroutine has stopped, so this clear is guaranteed to be the
	// last thing written to the spinner's line.
	s.write(clearLine)
}

// write is the single choke point for frame output, keeping it race-safe.
func (s *Spinner) write(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprint(s.writer, line)
}
