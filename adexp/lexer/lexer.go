/* Package lexer implements a lexer for ADEXP v3.1

## How does it work ?
	Given a ByteReader, you get a Lexer.
	Given the Lexer, you call ReadExpression() to get an expression.
*/
package lexer

import (
	"context"
	"io"
)

const (
	hyphen    byte = '-'
	separator byte = ' '
)

// An Expression codes for an expression
type Expression struct {
	Keyword []byte
	Value   []byte
}

// Lexer is what allows you to lex, and by streaming !
type Lexer struct {
	ctx    context.Context
	reader io.ByteReader
}

// Close to free up the optimisations in Lexer
func (lex *Lexer) Close() error {
	lex.reader = nil
	return nil
}

// New returns a new Lexer given a io.ByteReader
func (lex *Lexer) New(ctx context.Context, input io.ByteReader) *Lexer {
	return &lexer{ctx, input}
}
