/*
Package adexp provides marshalling and unmarshalling of ADEXP documents.
See: http://www.eurocontrol.int/sites/default/files/publication/files/20111001-adexp-spec-v3.1.pdf

Architecture:
	lexer provides the definition of a ADEXP lexer (DONE)
	parser provides the definition of a ADEXP parser (WIP)
	serialiser provides the definition of an ADEXP serialiser (TODO)
	this package wraps it all together so as to provide easy & simple unmarshalling of ADEXP documents (WIP)
*/
package adexp

// A Kind indicates of what concrete type the value is
type Kind uint8

// These are the three kinds defined for ADEXP
const (
	Err Kind = iota
	Primary
	Structured
	List
)

// value is the conterpart of a keyword
type value struct {
	kind  Kind
	value interface{}
}

// ADEXP is a Message in a ADEXP (Air traffic services Data EXchange Presentation) format.
type ADEXP map[string]value
