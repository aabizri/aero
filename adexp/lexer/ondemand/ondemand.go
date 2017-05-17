package ondemand

import (
	"io"
	"sync"
	"github.com/aabizri/aero/adexp/lexer"

	"github.com/pkg/errors"
)

// onDemandLexer is what allows you to lex, and by streaming !
type onDemandLexer struct {
	mu      sync.Mutex
	scanner io.RuneScanner
	state   stateFn
}

// New returns a new Lexer given a io.RuneScanner.
// It returns an on-demand Lexer that reads from the input as it is asked to lex.
// In the future, buffering Lexers may be build.
func New(input io.RuneScanner) lexer.Lexer {
	return &onDemandLexer{scanner: input, state: startState}
}

// Lex returns the next expression
func (odl *onDemandLexer) Lex() (*lexer.Lexeme, error) {
	// Update the next state function
	var (
		lexeme *lexer.Lexeme
		err    error
	)
	odl.mu.Lock()
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
	odl.mu.Unlock()

	return lexeme, err
}

// LexAll returns all the expressions
func (odl *onDemandLexer) LexAll() ([]lexer.Lexeme, error) {
	var results []lexer.Lexeme
	for i := 0; ; i++ {
		lexeme, err := odl.Lex()
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

func (odl *onDemandLexer) Close() error {
	odl.scanner = nil
	odl.state = nil
	return nil
}
