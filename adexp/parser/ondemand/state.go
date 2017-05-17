package ondemand

import (
	"io"

	"github.com/aabizri/aero/adexp/lexer"
	"github.com/aabizri/aero/adexp/parser"
	"github.com/pkg/errors"
)

type stateFn func(*onDemandParser) (*parser.Expression, stateFn, error)

// startState awaits a "TITLE" basic field.
func startState(odp *onDemandParser) (*parser.Expression, stateFn, error) {
	// Retrieve the next lexeme
	lex, err := odp.lexer.ReadLex()
	if err == io.EOF {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, errors.Wrap(err, "startState: error while retrieving next lexeme")
	}

	// If that lexeme isn't a keyword,  return an error
	if lex.Kind != lexer.LexemeKeyword {
		return nil, nil, errors.Wrapf(err, "startState: expected a keyword as first element, got a %s instead", lex.Kind.String())
	} else if lex.Value != parser.TITLEKeyword {
		return nil, nil, errors.Wrapf(err, "first field encontered is not \"%s\" but \"%s\"", parser.TITLEKeyword, lex.Value)
	}

	// Retrieve the value
	lex, err = odp.lexer.ReadLex()
	if err == io.EOF {
		return nil, nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, nil, errors.Wrap(err, "startState: error while retrieving next lexeme")
	}

	// Return
	expr := &parser.Expression{
		Kind:    parser.Primary,
		Keyword: parser.TITLEKeyword,
		Value:   parser.PrimaryField(lex.Value),
	}

	return expr, normalState, nil
}

// normalState is the normal state, i.e not currently in a field
func normalState(odp *onDemandParser) (*parser.Expression, stateFn, error) {
	// Retrieve the next lexeme
	lex, err := odp.lexer.ReadLex()
	if err == io.EOF {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, errors.Wrap(err, "normalState: error while retrieving next lexeme")
	}

	// If that lexeme isn't a keyword nor a BEGIN,  return an error
	if lex.Kind != lexer.LexemeKeyword && lex.Kind != lexer.LexemeBEGIN {
		return nil, nil, errors.Wrapf(err, "normalState: expected a keyword as first element, got a %s instead", lex.Kind.String())
	}

	// If we enconter a BEGIN, launch the beginState
	if lex.Kind == lexer.LexemeBEGIN {
		err := odp.lexer.UnreadLex()
		return nil, listState, err
	}

	// Else we encountered a keyword, so we launch a nonListState
	return nil, nonListState(lex.Value), nil
}

// nonListState returns a state where we expect a non-list field (i.e Primary or Sub fields)
func nonListState(keyword string) stateFn {
	return func(odp *onDemandParser) (*parser.Expression, stateFn, error) {
		// Retrieve the next lexeme
		lex, err := odp.lexer.ReadLex()
		if err == io.EOF {
			return nil, nil, err
		} else if err != nil {
			return nil, nil, errors.Wrap(err, "normalState: error while retrieving next lexeme")
		}

		// If that lexeme is neither a keyword nor a value, we return an error !
		if lex.Kind != lexer.LexemeKeyword && lex.Kind != lexer.LexemeValue {
			return nil, nil, errors.Errorf("nonListState: unexpected lexeme of kind \"%s\" instead of expected Keyword or Value", lex.Kind.String())
		}

		// Establish the expression
		expr := &parser.Expression{
			Keyword: keyword,
		}

		// If that lexeme is a keyword, then we have a subField
		// So we call parseSubField and return the returned value
		if lex.Kind == lexer.LexemeKeyword { /*
				err := odp.lexer.UnreadLex()
				if err != nil {
					return nil, nil, errors.Wrap(err, "nonListState: error while unreading last lexeme")
				}
				value, err := parseSubField(odp.lexer)
				if err != nil {
					return nil, nil, errors.Wrap(err, "nonListState: error in parseSubField")
				}
				expr.Value = value
			*/
		} else { // Else it's a basic field, so we assign it
			expr.Value = parser.PrimaryField(lex.Value)
		}

		return expr, normalState, nil
	}
}

// And parseSubField which starts off in a normalState, except we catch the expressions and put them in a parser.SubField
// THIS WILL NOT WORK UNTIL WE HAVE A LIST OF SUBFIELD-ENABLED KEYWORD AS WE NEED TO STOP AT SOME POINT
/*
func parseSubField(l lexer.LexScanner) (parser.StructuredField, error) {
	// Create a new parser
	p := New(l)
	values := make(parser.StructuredField)
	for i := 0; ; i++ {
		expr, err := p.Parse()
		if err != nil {
			return errors.Wrapf(err, "parseSubField: error in pass %d while parsing", i)
		}

	}
}*/

func parseListInternals(odp *onDemandParser) (parser.ListField, error) {
	// Create a new parser
	values := make(parser.ListField, 0)
	var state stateFn = normalState
	for i := 0; ; i++ {
		lex, err := odp.lexer.ReadLex()
		if err != nil {
			return nil, errors.Wrapf(err, "parseListInternals (pass #%d): error while retrieving next lexeme", i)
		}

		// We check if the next lexeme is an END keyword
		if lex.Kind == lexer.LexemeEND {
			odp.lexer.UnreadLex()
			break
		}

		// It isn't, so we unread and let NormalState manage it
		err = odp.lexer.UnreadLex()
		if err != nil {
			return nil, errors.Wrapf(err, "parseListInternals (pass %#d): error while unreading lexeme", i)
		}

		//Launch
		var expr *parser.Expression
		expr, state, err = state(odp)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrapf(err, "parseListInternals: (pass %#d): error while launching state", i)
		}

		if expr != nil {
			values = append(values, *expr)
		}
	}

	return values, nil
}

// In listState we expect a list, starting with BEGIN <keyword> [....] END <keyword>
func listState(odp *onDemandParser) (*parser.Expression, stateFn, error) {
	// Retrieve the next lexeme
	lex, err := odp.lexer.ReadLex()
	if err == io.EOF {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, errors.Wrap(err, "listState: error while retrieving next lexeme")
	}

	// If that lexeme isn't a BEGIN,  return an error
	if lex.Kind != lexer.LexemeBEGIN {
		return nil, nil, errors.Errorf("listState: expected BEGIN as first element, got a %s instead", lex.Kind.String())
	}

	// Retrieve the associated keyword
	lex, err = odp.lexer.ReadLex()
	if err == io.EOF {
		return nil, nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, nil, errors.Wrap(err, "listState: error while retrieving associated keyword")
	}

	// If that isn't a keyword, return an error
	if lex.Kind != lexer.LexemeKeyword {
		return nil, nil, errors.Errorf("listState: expected a keyword following a BEGIN, got a %s instead", lex.Kind.String())
	}

	// Create the expression
	expr := &parser.Expression{
		Kind:    parser.List,
		Keyword: lex.Value,
	}

	// Now we parse the internals of the list
	value, err := parseListInternals(odp)
	if err == io.EOF {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, errors.Wrapf(err, "listState: error while parsing list internals")
	}

	// We now expect an END
	lex, err = odp.lexer.ReadLex()
	if err == io.EOF {
		return nil, nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, nil, errors.Wrap(err, "listState: error while retrieving expected END lexeme")
	}

	if lex.Kind != lexer.LexemeEND {
		return nil, nil, errors.Errorf("listState: expected an END statement, got a %s instead", lex.Kind.String())
	}

	// And now a keyword
	lex, err = odp.lexer.ReadLex()
	if err == io.EOF {
		return nil, nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, nil, errors.Wrap(err, "listState: error while retrieving expected END lexeme")
	}

	// Check if this is a keyword
	if lex.Kind != lexer.LexemeKeyword {
		return nil, nil, errors.Errorf("listState: expected a keyword after an END lexeme, got a %s instead", lex.Kind.String())
	}

	// Check if the keyword is the same
	if lex.Value != expr.Keyword {
		return nil, nil, errors.Errorf("listState: list's associated keyword not consistent (BEGIN has %s , END has %s)", expr.Keyword, lex.Value)
	}

	// Return
	expr.Value = value
	return expr, normalState, nil
}
