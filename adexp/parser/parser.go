package parser

// A Kind is a keyword type
type Kind uint8

/* These are the Kinds: keyword types
In ADEXP you can have three keyword types:
	- Primary Field (e.g "-TITLE I LOVE ROCK N ROLL")
	- Sub-Field (e.g "-GEO -GEOID 01 -LATTD 520000N -LONGTD 0150000W")
	- List Field (e.g "-BEGIN ADDR -FAC LLEVZPZX -FAC LFFFZQZX -END ADDR")
*/
const (
	Primary Kind = iota
	Structured
	List
)

func (k Kind) String() string {
	switch k {
	case Primary:
		return "primary field"
	case Structured:
		return "structured (aka sub) field"
	case List:
		return "list field"
	default:
		return "unknown kind"
	}
}

const (
	TITLEKeyword = "TITLE"
)

// A Value is the information-carrying part of an expression
// It can take one of three kinds
type Value interface{}

type (
	// A PrimaryField has only a text value
	PrimaryField string
	// A StructuredField has several subfields, but it isn't a list
	StructuredField map[string]Expression
	// A ListField is a list of subfields
	ListField []Expression
)

// An Expression is composed of a keyword and of a value
type Expression struct {
	Kind    Kind
	Keyword string
	Value   Value
}

// Parser is the interface for types implementing Parse
type Parser interface {
	// Parse returns the next top-level expression in ADEXP v3.1
	Parse() (*Expression, error)
}
