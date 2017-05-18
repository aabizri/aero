/*Package scannify provides methods to convert a lexer.LexReader to a lexer.LexScanner*/
package scannify

import (
	"errors"
	"sync"

	"github.com/aabizri/aero/adexp/lexer"
)

// lexScanCloser is the implementation of a lexer.LexScanCloser by wrapping a lexer.LexReader
type lexScanCloser struct {
	mu       sync.Mutex
	reader   lexer.LexReader
	previous *lexer.Lexeme
	unreaded bool
}

// New takes a lexer.LexReader and return a wrapper implementing lexer.LexScanner
//
// Warning: closing the returned lexer.LexScanCloser doesn't close the embedded reader
func New(reader lexer.LexReader) lexer.LexScanCloser {
	return &lexScanCloser{reader: reader}
}

func (ls *lexScanCloser) ReadLex() (*lexer.Lexeme, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.unreaded {
		ls.unreaded = false
		return ls.previous, nil
	}
	lex, err := ls.reader.ReadLex()
	ls.previous = lex
	return lex, err
}

func (ls *lexScanCloser) UnreadLex() error {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.unreaded {
		return errors.New("UnreadLex: cannot unread more than once")
	}
	ls.unreaded = true
	return nil
}

func (ls *lexScanCloser) Close() error {
	ls.mu.Lock()
	ls.previous = nil
	ls.mu.Unlock()
	return nil
}
