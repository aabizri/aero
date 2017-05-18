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

// In startState we expect either separators or a hyphen.
// Separators are ignored here, we continue on until we find a hyphen or EOF.
// Hyphens aren't ignored, they indicate a new keyword, in that case we return a keywordState
func startState(odl *onDemandLexReader) (*lexer.Lexeme, stateFn, error) {
	for i := 0; ; i++ {
		// Get the next byte
		current, _, err := odl.scanner.ReadRune()
		if err == io.EOF { // EOFs are completely legal before any start of anything. They just mean that we have an empty input.
			return nil, nil, err
		} else if err != nil {
			return nil, nil, errors.Wrapf(err, "valueState (iteration %d): error while reading next rune", i)
		}

		switch {
		// In case we enconter a hyphen, then we launch into keywordState
		case current == hyphen:
			return nil, keywordState, nil

		// If we are still seing separators, we continue on
		case unicode.IsSpace(current):

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
		switch err {
		case io.EOF:
			return nil, nil, io.ErrUnexpectedEOF
		case nil:
		default:
			return nil, nil, errors.Wrapf(err, "keywordState (iteration %d): error while reading next rune", i)
		}

		// Switch
		switch {
		// A keyword can only be composed of upper-case characters
		case unicode.IsUpper(current):
			runes = append(runes, current)
			inKeyword = true

		// If we've encontered a separator after having first encountered proper text, we break
		case unicode.IsSpace(current) && inKeyword:
			break Loop

		// In case we still haven't encontered the first character, we continue on
		case unicode.IsSpace(current):
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
		// Get the rune
		current, _, err := odl.scanner.ReadRune()
		switch err {

		// If we encounter an EOF here, it means a keyword has no associated value or other keywords, which is invalid.
		// As such, we return io.ErrUnexpectedEOF
		case io.EOF:
			return nil, nil, io.ErrUnexpectedEOF

		case nil:
		default:
			return nil, nil, errors.Wrapf(err, "postKeywordState (iteration %d): error while reading next rune", i)
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
		lastNonSep int                                       // index of the latest non-separator valid character
	)
Loop:
	for i := 0; ; i++ {
		// Get the rune
		current, _, err := odl.scanner.ReadRune()

		// If we get an EOF in the value, it is absolutely normal except if we encontered no previous non-separator values, so we simply stop the loop and return what we have
		switch {
		case err == io.EOF && lastNonSep != 0:
			break Loop
		case err == io.EOF:
			return nil, nil, io.ErrUnexpectedEOF
		case err != nil:
			return nil, nil, errors.Wrapf(err, "valueState (iteration %d): error while reading next rune", i)
		}

		// Switch
		switch {
		// A value can be composed of upper-case letters and/or digits, as well as separators
		// We note the position of the last non-separator element so that we remove trailing separators when we enconter a new keyword
		case unicode.IsUpper(current) || unicode.IsDigit(current):
			lastNonSep = i
			runes = append(runes, current)

		// Separators can be valid inside a value when they are surrounded by upper-case letters and/or digits.
		// Here we append them to the slice but we will slice later to remove the trailing separators.
		case unicode.IsSpace(current):
			runes = append(runes, current)

		// A value is terminated by either a new keyword or EOF, we checked for EOF previously, here we check for a new keyword, indicated by a hyphen.
		case current == hyphen:
			break Loop

		default:
			return nil, nil, errors.Errorf("valueState: unexpected character at offset %d: \"%s\"", i, string(current))
		}
	}

	// We slice to remove the trailing separators
	str := string(runes[:lastNonSep+1])
	lexeme := &lexer.Lexeme{
		Kind:  lexer.LexemeValue,
		Value: str,
	}

	return lexeme, keywordState, nil
}

// in postListBoundState, we expect an alphanumeric value of lexer.LexemeKeyword kind
// and we return a startField or EOF
func postListBoundState(odl *onDemandLexReader) (*lexer.Lexeme, stateFn, error) {

	var runes = make([]rune, 0, expectedMaxKeywordLength) // we expect a max keyword length, this shaves off time in growing the slice

Loop:
	for i := 0; ; i++ {
		// Get the rune
		current, _, err := odl.scanner.ReadRune()
		switch {
		case err == io.EOF && len(runes) == 0: // Here, an EOF is illegal if we haven't yet encontered a keyword, we thus return an io.ErrUnexpectedEOF
			return nil, nil, io.ErrUnexpectedEOF
		case err == io.EOF:
			break Loop
		case err != nil:
			return nil, nil, errors.Wrapf(err, "postListBoundState (iteration %d): error while reading next rune", i)
		}

		// Switch
		switch {
		// A keyword is only upper-case
		case unicode.IsUpper(current):
			runes = append(runes, current)

		// A keyword is composed of solely one word
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
