// Package buffering implements a buffering lexer of ADEXP v3.1 format.
package buffering

import (
	"io"

	"github.com/aabizri/aero/adexp/lexer"

	"github.com/pkg/errors"
)

type bufferingLexer struct {
	// The mbedded Reader
	lexer.LexReader

	// Buffer holds the lexemes coming up
	buffer chan *lexer.Lexeme

	// Backup holds the previous lexers
	backup   chan *lexer.Lexeme
	unreaded bool

	// Done signals the end of the background goroutine
	done chan struct{}

	// Err is the error
	err error
}

// New returns a buffering LexScanCloser.
// It will accumulate lexemes until closed.
func New(l lexer.LexReader, bufferLen int) lexer.LexScanCloser {
	bl := &bufferingLexer{
		LexReader: l,
		buffer:    make(chan *lexer.Lexeme, bufferLen),
		backup:    make(chan *lexer.Lexeme, 1),
		done:      make(chan struct{}, 1),
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
			// It doesn't, so let's continue
			lexeme, err := bl.LexReader.ReadLex()
			if err != nil {
				bl.err = err
				break Loop
			}

			// Put the lexeme in the buffer
			bl.buffer <- lexeme
		}

	}

	// Close the buffer as we've either touched EOF, encontered an error, or been told to stop
	close(bl.buffer)
}

// Close closes the lexer
// If the provided lexer implements io.Closer, then it is called as well.
func (bl *bufferingLexer) Close() error {
	// If the buffer channel is not closed we send a signal
	if _, ok := <-bl.buffer; ok {
		bl.done <- struct{}{}
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
	var lexeme *lexer.Lexeme
	// If we don't unread, then get a new lexeme and push it to the backup buffer
	// Popping out the previous one if there is one
	if bl.unreaded == false {
		// If bl.buffer is closed, the we return nil, err
		l, ok := <-bl.buffer
		if !ok {
			return nil, bl.err
		}

		// Pop out and push in
		if len(bl.backup) == 1 { // If the previous operation was an unreaded read.
			<-bl.backup
		}
		bl.backup <- l
		lexeme = l
	} else { // If we do unread, pop out the lexeme from the buffer and return it
		if len(bl.backup) == 0 {
			return nil, errors.New("backup buffer is empty, should not happen! ")
		}
		l := <-bl.backup
		bl.backup <- l
		lexeme = l
		bl.unreaded = false
	}

	// Now, we return the lexeme
	return lexeme, nil
}

// UnreadLex unreads a lexeme
func (bl *bufferingLexer) UnreadLex() error {
	bl.unreaded = true
	return nil
}
