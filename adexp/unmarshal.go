package adexp

import (
	"bytes"
	"io"

	"github.com/aabizri/aero/adexp/parser"
)

// UnmarshalText unmarshals the given text to msg.
// Note that it is much more efficient to use Decoder with a streaming io.Reader.
func (msg ADEXP) UnmarshalText(text []byte) error {
	r := bytes.NewReader(text)
	enc := NewDecoder(r)
	err := enc.Decode(msg)
	return err
}

// A Decoder decodes an input stream to an ADEXP map
type Decoder struct {
	reader     io.Reader
	parserFunc func(io.Reader) parser.Parser
}

// NewDecoder returns a default Decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

// SetParser sets the Parser-obtaining function to be used.
// If used after the first call to Decode, this results in an error.
func (dec *Decoder) SetParser(new func(io.Reader) parser.Parser) error {
	dec.parserFunc = new
	return nil
}

// Decode decodes the input stream to the given ADEXP msg
func (dec *Decoder) Decode(msg ADEXP) error {
	// Do something
	return nil
}
