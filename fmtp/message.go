package fmtp

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// A Message is composed of a Header and a Body
type Message struct {
	header *header
	Body   io.Reader

	// How much has already been written
	written int
}

// WriteTo writes a Message to the given io.Writer
// If m.Body has a Len or Bytes method, then it is used, if it doesn't the request is copied to a buffer before writing it over.
// This consumes the Message Body.
func (m *Message) WriteTo(w io.Writer) (int64, error) {
	// Check if message is valid
	if m.header == nil {
		return 0, errors.New("WriteTo: cannot write message as header is nil")
	}

	// Establish the interfaces
	type lener interface {
		Len() int
	}
	type byteser interface {
		Bytes() int
	}

	// If the header has a set content-length, simply use it. If it doesn't, extract it from reader.
	if m.header.length == 0 {
		fmt.Println("Message has no set length, creating it")
		switch r := m.Body.(type) {
		case lener:
			m.header.length = uint16(r.Len())
		case byteser:
			m.header.length = uint16(r.Bytes())
		default:
			fmt.Println("Copying io.Reader into memory")
			buf := &bytes.Buffer{}
			n, err := io.Copy(buf, m.Body)
			if err != nil {
				return 0, errors.Wrap(err, "Read: error while retrieving size of reader")
			}
			m.header.length = uint16(n)
			m.Body = buf
		}
	}

	// Now write the header to the writer
	n1, err1 := m.header.WriteTo(w)
	if err1 != nil && err1 != io.EOF {
		return n1, errors.Wrap(err1, "Read: error while reading header")
	}

	// And read from body
	n2, err2 := io.Copy(w, m.Body)
	if err2 != nil && err2 != io.EOF {
		return n1 + n2, errors.Wrap(err2, "Read: error while reading body")
	}

	return n1 + n2, nil
}

// ReadFrom creates a m.Message from an io.Reader.
func (m *Message) ReadFrom(r io.Reader) (int64, error) {
	// First we decode the header
	h := &header{}
	n1, err := h.ReadFrom(r)
	if err != nil {
		return n1, err
	}
	m.header = h

	// Now, given the header-indicated size we create a buffer of that size
	fmt.Printf("length: %d, n: %d\n", h.length, n1)
	content := make([]byte, h.bodyLen())
	n2, err := r.Read(content)
	if n2 != h.bodyLen() {
		return n1 + int64(n2), errors.New("Read: reader read less than the amount noted in the header: ILLEGAL")
	}

	// And we create a bufio.Reader from it
	body := bytes.NewReader(content)
	m.Body = body

	return n1 + int64(n2), nil
}

// NewOperatorMessageFromString returns a message of Operator type
func NewOperatorMessageFromString(txt string) (*Message, error) {
	r := strings.NewReader(txt)
	msg, err := NewMessage(Operator, r)
	if err != nil {
		return msg, err
	}
	msg.header.setLength(uint16(len(txt)))
	return msg, nil
}

// NewMessage returns a message of either Operational or Operator type
func NewMessage(typ uint8, r io.Reader) (*Message, error) {
	return &Message{
		header: newHeader(typ),
		Body:   r,
	}, nil
}

// NewOperationalMessage returns a message of Operational type
func NewOperationalMessage(r io.Reader) (*Message, error) {
	return NewMessage(Operational, r)
}

// NewOperatorMessage returns a message of Operator type
func NewOperatorMessage(r io.Reader) (*Message, error) {
	return NewMessage(Operator, r)
}

// newIDRequestMessage returns an identification request message
func newIDRequestMessage(sender ID, receiver ID) (*Message, error) {
	idr, err := newidRequest(sender, receiver)
	if err != nil {
		return nil, errors.Wrap(err, "newIDRequestMessage: error while creating new IDRequest message")
	}
	r, err := idr.Reader()
	if err != nil {
		return nil, errors.Wrap(err, "newIDRequestMessage: error when recovering reader for id request")
	}

	return NewMessage(identification, r)
}

// newIDResponseMessage returns an identification response message
func newIDResponseMessage(accept bool) (*Message, error) {
	idr := newidResponse(accept)
	r, err := idr.Reader()
	if err != nil {
		return nil, errors.Wrap(err, "newIDResponseMessage: error when recovering reader for id response")
	}
	return NewMessage(identification, r)
}

// newSystemMessage returns a system message
func newSystemMessage(ss systemSig) (*Message, error) {
	return NewMessage(identification, bytes.NewReader(ss[:]))
}
