/*
Package lexer defines a lexer for ADEXP v3.1

You can find one implementation in the ondemand subpackage
*/
package lexer

// A Kind indicates the kind of the lexeme
type Kind uint8

// String implements Stringer
func (k Kind) String() string {
	switch k {
	case LexemeKeyword:
		return "keyword"
	case LexemeValue:
		return "value"
	case LexemeBEGIN:
		return "BEGIN"
	case LexemeEND:
		return "END"
	default:
		return "unknown"
	}
}

// Those are the kinds
const (
	LexemeKeyword Kind = iota // A keyword is what is preceded by a '-' or a 'BEGIN' or an 'END'
	LexemeValue               // A value is what isn't preceded by a by a '-' or a 'BEGIN' or an 'END'

	LexemeBEGIN
	LexemeEND
)

// A Lexeme holds a lexeme, i.e an expression tokenised by the lexer.
// It is composed of a Kind (i.e what is the type of that lexeme) and a Value
type Lexeme struct {
	Kind  Kind
	Value string
}

// The Lexer interface allows you to read expressions
type Lexer interface {
	// Lex should return io.EOF when no more lexemes are available
	Lex() (*Lexeme, error)
}

// The LexerAll interface allows you to read all expression
type LexerAll interface {
	LexerAll() ([]Lexeme, error)
}
