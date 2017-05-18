// DONE: Optimise rune slice allocation

package ondemand

import (
	"io"
	"unicode"

	"github.com/aabizri/aero/adexp/lexer"

	"github.com/pkg/errors"
)

const (
	hyphen                   rune = '-'
	expectedMaxKeywordLength      = 9
	expectedMaxValueLength        = 12
)

// stateFn represents the state of the Lexer as a function that returns the next state
type stateFn func(*onDemandLexReader) (*lexer.Lexeme, stateFn, error)

// in startState we expect either separators or a hyphen or a BEGIN.
// separators are ignored here
// hyphens aren't, they indicate a new keyword, in that case we return a keywordState
// in case of a B for BEGIN we return a beginState
func startState(odl *onDemandLexReader) (*lexer.Lexeme, stateFn, error) {
	for i := 0; ; i++ {
		// Get the next byte
		current, _, err := odl.scanner.ReadRune()
		if err != nil {
			return nil, nil, err
		}

		// Switch for the three cases
		switch {
		// In this case return the hyphen lexeme
		case current == hyphen:
			return nil, keywordState, nil
		case unicode.IsSpace(current): // We let it run
		default:
			return nil, nil, errors.Errorf("startState: unexpected character at offset %d: \"%s\"", i, string(current))
		}
	}
}

// in keywordState we return a keyword until the next separator
// we'll return either a lexeme in a lexer.LexemeKeyword, lexer.BeginKeyword or lexer.EndKeyword
func keywordState(odl *onDemandLexReader) (*lexer.Lexeme, stateFn, error) {
	var (
		runes     = make([]rune, 0, 9) // we expect a max keyword length, this shaves off time in growing the slice
		inKeyword bool
	)
Loop:
	for i := 0; ; i++ {
		// Get the current byte
		current, _, err := odl.scanner.ReadRune()
		if err != nil {
			return nil, nil, err
		}

		// Switch
		switch {
		case unicode.IsUpper(current): // We only accept uppercase keywords
			runes = append(runes, current)
			inKeyword = true
		case unicode.IsSpace(current) && inKeyword: // If we've encontered a separator after having first encountered proper text, we break
			break Loop
		case unicode.IsSpace(current): // We just continue until the start of the keyword then
		default:
			return nil, nil, errors.Errorf("keywordState: unexpected character at offset %d: \"%s\"", i, string(current))
		}
	}
	kind := lexer.LexemeKeyword
	nextState := postKeywordState

	str := string(runes)
	// If the keyword is BEGIN, then we start a postListBoundState
	if str == "BEGIN" {
		kind = lexer.LexemeBEGIN
		nextState = postListBoundState
	}
	// If the keyword is END, then we start a postListBoundState
	if str == "END" {
		kind = lexer.LexemeEND
		nextState = postListBoundState
	}

	lexeme := &lexer.Lexeme{
		Kind:  kind,
		Value: str,
	}

	return lexeme, nextState, nil
}

// in postKeywordState we expect either
// 	a hyphen signaling a new keyword, and as such we're entering a subFieldState
// 	an alphanumeric text signaling a value, and as such we're entering a basicFieldState
func postKeywordState(odl *onDemandLexReader) (*lexer.Lexeme, stateFn, error) {
	for i := 0; ; i++ {
		// Get the next byte
		current, _, err := odl.scanner.ReadRune()
		if err != nil {
			return nil, nil, err
		}

		// Switch for the three cases
		switch {
		// If we enconter a hyphen, then we have a subfield, and so we return a keywordState
		case current == hyphen:
			return nil, keywordState, nil

		// If we get a letter or digit after the separator, then we have a basic field.
		// So we return a valueState.
		case unicode.IsUpper(current) || unicode.IsDigit(current):
			odl.scanner.UnreadRune()
			return nil, valueState, nil

		// We ignore separators
		case unicode.IsSpace(current): // We let it run

		// If we have an unknown character, its an error !
		default:
			return nil, nil, errors.Errorf("postKeywordState: unexpected character at offset %d: \"%s\"", i, string(current))
		}
	}
}

// in valueState we expect only alphanumeric values, so if we encounter a hyphen, we return a keywordState
func valueState(odl *onDemandLexReader) (*lexer.Lexeme, stateFn, error) {
	var (
		runes      = make([]rune, 0, expectedMaxValueLength) // we expect a max value length, this shaves off time in growing the slice
		lastNonSep int
		nextState  stateFn
	)
Loop:
	for i := 0; ; i++ {
		// Get the current byte
		current, _, err := odl.scanner.ReadRune()
		// If we get an IOF in the value, then we simply stop the loop and return what we have
		if err == io.EOF && i != 0 {
			break
		} else if err != nil {
			return nil, nil, err
		}

		// Switch
		switch {
		case unicode.IsUpper(current) || unicode.IsDigit(current):
			lastNonSep = i
			runes = append(runes, current)
		case unicode.IsSpace(current):
			runes = append(runes, current)
		case current == hyphen:
			nextState = keywordState
			break Loop
		default:
			return nil, nil, errors.Errorf("valueState: unexpected character at offset %d: \"%s\"", i, string(current))
		}
	}

	str := string(runes)[:lastNonSep+1]
	lexeme := &lexer.Lexeme{
		Kind:  lexer.LexemeValue,
		Value: str,
	}

	return lexeme, nextState, nil
}

// in postListBoundState, we expect an alphanumeric value of lexer.LexemeKeyword kind
// and we return a startField or EOF
func postListBoundState(odl *onDemandLexReader) (*lexer.Lexeme, stateFn, error) {

	var runes = make([]rune, 0, expectedMaxKeywordLength) // we expect a max keyword length, this shaves off time in growing the slice
Loop:
	for i := 0; ; i++ {
		// Get the current byte
		current, _, err := odl.scanner.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}

		// Switch
		switch {
		case unicode.IsUpper(current):
			runes = append(runes, current)
		case unicode.IsSpace(current):
			break Loop
		default:
			return nil, nil, errors.Errorf("postListBoundState: unexpected character at offset %d: \"%s\"", i, string(current))
		}
	}

	str := string(runes)
	lexeme := &lexer.Lexeme{
		Kind:  lexer.LexemeKeyword,
		Value: str,
	}

	return lexeme, startState, nil
}
