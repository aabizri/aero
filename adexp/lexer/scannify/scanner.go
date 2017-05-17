package scannify

import (
	"errors"
	"sync"

	"github.com/aabizri/aero/adexp/lexer"
)

type LexScanner struct {
	mu       sync.Mutex
	reader   lexer.LexReader
	previous *lexer.Lexeme
	unreaded bool
}

func New(reader lexer.LexReader) lexer.LexScanner {
	return &LexScanner{reader: reader}
}

func (ls *LexScanner) ReadLex() (*lexer.Lexeme, error) {
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

func (ls *LexScanner) UnreadLex() error {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.unreaded {
		return errors.New("UnreadLex: cannot unread more than once")
	}
	ls.unreaded = true
	return nil
}

func (ls *LexScanner) Close() error {
	ls.mu.Lock()
	ls.previous = nil
	ls.mu.Unlock()
	return nil
}
