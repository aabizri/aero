package lexer

import (
	"context"
	"io"

	"github.com/pkg/errors"
)

const (
	hyphen    byte = '-'
	separator byte = ' '
)

// onDemandLexer is what allows you to lex, and by streaming !
type onDemandLexer struct {
	scanner io.ByteScanner
}

// NewOnDemand returns a new Lexer given a io.ByteScanner.
// It returns an on-demand Lexer that reads from the input as it is asked to lex.
// In the future, buffering Lexers may be build.
func NewOnDemand(input io.ByteScanner) Lexer {
	return &onDemandLexer{scanner: input}
}

// ReadExpression returns the next expression
// If that's the last Expression, you get a io.EOF as error
func (lex *onDemandLexer) ReadExpression(ctx context.Context) (*Expression, error) {
	var (
		cursorInKeyword  bool
		keywordIndex     int
		keywordLen       int
		valueIndex       int
		valueLen         int
		bytes            []byte
		cursorInValue    bool
		lastNonSeparator int
	)

	// Walk
Loop:
	for i := 0; ; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Set up the bytes
		b, err := lex.scanner.ReadByte()
		switch {
		case err == io.EOF && !cursorInValue:
			return nil, err
		case err != nil && err != io.EOF:
			return nil, errors.Wrapf(err, "adexp/lexer.ReadExpression: unexpected read error at offset %d of this expression", i)
		}
		bytes = append(bytes, b)

		// Check for last non-separator
		if b != separator && b != hyphen {
			lastNonSeparator = i
		}

		// Now we actually walk
		switch {
		// If we got a hyphen or io.EOF while parsing the value, then we have EOF
		case (b == hyphen || err == io.EOF) && cursorInValue:
			valueLen = lastNonSeparator - valueIndex
			// If we didn't finish the scanner, then we unread
			if err != io.EOF {
				err := lex.scanner.UnreadByte()
				if err != nil {
					return nil, errors.Wrapf(err, "adexp/lexer: error while unreading byte with offset %d in this expression", i)
				}
			}
			break Loop

		// If we got a hyphen as first element, then we have a start !
		case b == hyphen && !cursorInKeyword && !cursorInValue:
			// Mark the fact we are currently in a keyword
			cursorInKeyword = true

			// Mark the beginning of the keyword
			keywordIndex = i + 1

		// If we have a separator, and we're parsing the keyword, then we've passed the keyword
		case b == separator && cursorInKeyword:
			cursorInKeyword = false
			keywordLen = lastNonSeparator - keywordIndex

		// If we just passed a separator, and we've touched a non-separator character, then we have the beginning of a value
		case b != separator && b != hyphen && !cursorInKeyword && !cursorInValue:
			cursorInValue = true
			valueIndex = i
		}

	}

	// Now split the bytes into the the keyword and value
	expression := &Expression{
		Keyword: bytes[keywordIndex : keywordIndex+keywordLen+1],
		Value:   bytes[valueIndex : valueIndex+valueLen+1],
	}
	return expression, nil
}

// Close to free up the optimisations in Lexer
func (lex *onDemandLexer) Close() error {
	lex.scanner = nil
	return nil
}
