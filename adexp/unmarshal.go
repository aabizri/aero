package adexp

import (
	"bytes"
	"io"
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
	reader io.Reader
}

// NewDecoder returns a default Decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

// Decode decodes the input stream to the given ADEXP msg
func (dec *Decoder) Decode(msg ADEXP) error {
	// Do something
	return nil
}
