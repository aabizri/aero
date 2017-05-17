// Package buffering implements a buffering lexer of ADEXP v3.1 format.
package buffering

import (
	"io"

	"github.com/aabizri/aero/adexp/lexer"

	"github.com/pkg/errors"
)

type bufferingLexer struct {
	lexer.LexReader
	buffer chan *lexer.Lexeme
	done   chan struct{}
	err    error
}

// NewLexReader returns a buffering LexReader.
// It will accumulate lexemes until closed.
func NewLexReader(l lexer.LexReader, bufferLen int) lexer.LexReader {
	bl := &bufferingLexer{
		LexReader: l,
		buffer:    make(chan *lexer.Lexeme, bufferLen),
		done:      make(chan struct{}),
	}

	go bl.background()

	return bl
}

// It is background() that does the job
func (bl *bufferingLexer) background() {
Loop:
	for {
		// Check if the done channel has anything to give us
		select {
		case <-bl.done:
			break Loop
		default:
		}

		// It doesn't, so let's continue
		lexeme, err := bl.LexReader.ReadLex()
		if err != nil {
			bl.err = err
			break
		}

		// Put the lexeme in the buffer
		bl.buffer <- lexeme
	}

	// Close the buffer as we've either touched EOF, encontered an error, or been told to stop
	close(bl.buffer)
}

// Close closes the lexer
// If the provided lexer implements io.Closer, then it is called as well.
func (bl *bufferingLexer) Close() error {
	// We send the done signal
	bl.done <- struct{}{}

	// We empty the buffer, waiting for the buffer to be closed
	for {
		_, ok := <-bl.buffer
		if !ok {
			break
		}
	}

	// We close the embedded lexer if it supports closing
	if c, ok := bl.LexReader.(io.Closer); ok {
		err := c.Close()
		return errors.Wrap(err, "error when closing provided lexer")
	}

	return nil
}

// Lex returns the next expression
func (bl *bufferingLexer) ReadLex() (*lexer.Lexeme, error) {
	// If bl.buffer is closed, the we return nil, err
	lexeme, ok := <-bl.buffer
	if !ok {
		return nil, bl.err
	}

	// Else, we return the lexeme
	return lexeme, nil
}
