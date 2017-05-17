package ondemand

import (
	"io"
	"sync"

	"github.com/aabizri/aero/adexp/lexer"

	"github.com/pkg/errors"
)

// onDemandLexReader is what allows you to lex, and by streaming !
type onDemandLexReader struct {
	mu      sync.Mutex
	scanner io.RuneScanner
	state   stateFn
}

// NewLexReader returns a new LexReader given a io.RuneScanner.
// It returns an on-demand LexReader that reads from the input as it is asked to lex.
func NewLexReader(input io.RuneScanner) lexer.LexReader {
	return &onDemandLexReader{scanner: input, state: startState}
}

// Lex returns the next expression
func (odl *onDemandLexReader) ReadLex() (*lexer.Lexeme, error) {
	// Update the next state function
	var (
		lexeme *lexer.Lexeme
		err    error
	)
	odl.mu.Lock()
	defer odl.mu.Unlock()
	for i := 0; lexeme == nil; i++ {
		lexeme, odl.state, err = odl.state(odl)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrapf(err, "Lex: lexing error in iteration %d", i)
		}

		if odl.state == nil {
			err = io.EOF
		}
	}

	return lexeme, err
}

// LexAll returns all the expressions
func (odl *onDemandLexReader) LexAll() ([]lexer.Lexeme, error) {
	var results []lexer.Lexeme
	for i := 0; ; i++ {
		lexeme, err := odl.ReadLex()
		switch err {
		case io.EOF:
			return results, nil
		case nil:
		default:
			return results, errors.Wrapf(err, "LexAll: error in lexing lexeme #%d", i)
		}
		results = append(results, *lexeme)
	}
}

func (odl *onDemandLexReader) Close() error {
	odl.mu.Lock()
	odl.scanner = nil
	odl.state = nil
	odl.mu.Unlock()
	return nil
}
