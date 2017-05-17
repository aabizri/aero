// Package ondemand of the parser subpackage implements an on-demand parser for ADEXP V3.1
package ondemand

import (
	"io"
	"sync"

	"github.com/aabizri/aero/adexp/lexer"
	"github.com/aabizri/aero/adexp/parser"
	"github.com/pkg/errors"
)

type onDemandParser struct {
	mu    sync.Mutex
	lexer lexer.LexScanner
	state stateFn
}

// New creates a new on-demand parser.Parser
func New(lex lexer.LexScanner) parser.Parser {
	return &onDemandParser{
		lexer: lex,
		state: startState,
	}
}

// Parse returns the next expression
func (odp *onDemandParser) Parse() (*parser.Expression, error) {
	var (
		expr *parser.Expression
		err  error
	)
	odp.mu.Lock()
	defer odp.mu.Unlock()
	for i := 0; expr == nil; i++ {
		expr, odp.state, err = odp.state(odp)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrapf(err, "Parse (pass %#d): error while parsing next expression", i)
		}

		// If we have no state left, we return
		if odp.state == nil {
			err = io.EOF
		}
	}
	return expr, err
}

// Close closes an onDemandParser.
// It does NOT close the included lexer.
func (odp *onDemandParser) Close() error {
	odp.mu.Lock()
	odp.state = nil
	odp.mu.Unlock()
	return nil
}
