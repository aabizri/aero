package adexp

import (
	"bytes"
	"io"

	"github.com/aabizri/aero/adexp/parser"
	"github.com/pkg/errors"
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
	parser     parser.Parser

	started bool
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
	if dec.started {
		return errors.New("cannot set parser to be used when decoding has already been started")
	}
	dec.parserFunc = new
	return nil
}

// Decode decodes the input stream to the given ADEXP msg
func (dec *Decoder) Decode(msg ADEXP) error {
	// Note that we have started
	dec.started = true

	// Build the parser and remove the other fields
	dec.parser = dec.parserFunc(dec.reader)
	dec.reader = nil
	dec.parserFunc = nil

	// Now we parse
Loop:
	for i := 0; ; i++ {
		expr, err := dec.parser.Parse()

		// Check for parsing errors
		switch err {
		// If we get an EOF, we quit
		case io.EOF:
			break Loop
		case nil:
		// If we have an unexpected error
		default:
			return errors.Wrapf(err, "Decode (expression %d): parsing error", i)
		}

		// Check that expr isn't nil, it shouldn't !
		if expr == nil {
			return errors.Errorf("Decode (expression %d): we got an unexpected nil expression", i)
		}

		// Now apply that to our map
		switch expr.Kind {
		case parser.Primary:
			if pf, ok := expr.Value.(parser.PrimaryField); ok {
				val := value{
					kind:  Primary,
					value: string(pf),
				}
				msg[expr.Keyword] = val
			} else {
				return errors.Errorf("Decode (expression %d): parser indicated kind %s but it doesn't match with value (%T)", i, expr.Kind, expr)
			}
		// TODO:
		case parser.Structured:
		case parser.List:
		}
	}

	// And finished !
	return nil
}
