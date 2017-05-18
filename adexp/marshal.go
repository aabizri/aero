package adexp

import (
	"bytes"
	"io"
)

// MarshalText marshals an ADEXP message to string
func (msg ADEXP) MarshalText() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := NewEncoder(buf)
	err := enc.Encode(msg)
	return buf.Bytes(), err
}

// An Encoder writes an ADEXP map to an output stream
// Note that it is much more efficient to use Encoder with a streaming io.Writer.
type Encoder struct {
	sep    string
	indent string
	writer io.Writer
}

// NewEncoder returns a new encoder that writes to w
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		sep:    " ",
		indent: "\t",
		writer: w,
	}
}

// SetIndent sets the indentation. If indent is "", then no indentation will be applied.
// The default indentation is \t. If you want tabs as a separator then you might use SetSeparator.
func (enc *Encoder) SetIndent(indent string) error {
	enc.indent = indent
	return nil
}

// SetSep sets the separator. There has to be at least one width of separator.
func (enc *Encoder) SetSep(sep string) error {
	enc.sep = sep
	return nil
}

// Encode encodes a given ADEXP message
func (enc *Encoder) Encode(msg ADEXP) error {
	// DO SOMETHING
	return nil
}
