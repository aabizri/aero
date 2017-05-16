package scannify

import (
	"errors"
	"io"
	"sync"
	"unicode/utf8"
)

// RuneScanner is the interface that adds the UnreadRune method to the basic ReadRune method.
type RuneScanner struct {
	reader   io.RuneReader
	mu       sync.Mutex
	previous rune
	unreaded bool
}

// ReadRune reads a single UTF-8 encoded Unicode character and returns the rune and its size in bytes.
// If no character is available, err will be set.
func (rs *RuneScanner) ReadRune() (rune, int, error) {
	rs.mu.Lock()
	switch rs.unreaded {
	case false:
		r, size, err := rs.reader.ReadRune()
		rs.previous = r
		if err != nil {
			return r, size, err
		}
		return r, size, nil
	case true:
		rs.unreaded = false
		return rs.previous, utf8.RuneLen(rs.previous), nil
	}
	rs.mu.Unlock()
	return 0, 0, nil
}

// UnreadRune causes the next call to ReadRune to return the same rune as the previous call to ReadRune. It is an error to call UnreadRune twice without an intervening call to ReadRune.
func (rs *RuneScanner) UnreadRune() error {
	rs.mu.Lock()
	if rs.unreaded {
		return errors.New("Scanner: cannot unread more than once")
	}

	rs.unreaded = true
	rs.mu.Unlock()
	return nil
}
