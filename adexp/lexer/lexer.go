/*
Package lexer implements a lexer for ADEXP v3.1

The implementation in this package accepts a io.ByteScanner, but there's obviously other ways to do that (chan of Byte, etc.)
It is geared toward use in a pipeline, that is giving the Lexer to a function that will read as will.
*/
package lexer

import (
	"context"
)

// An Expression codes for an expression
type Expression struct {
	Keyword []byte
	Value   []byte
}

// The Lexer interface allows you to read expressions
type Lexer interface {
	Close() error
	ReadExpression(ctx context.Context) (*Expression, error)
}
