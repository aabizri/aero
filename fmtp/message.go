package fmtp

import (
	"bytes"
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

// buffer switches the Body of a message for a buffer, returning the size read
func (msg *Message) buffer() (uint16, error) {
	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, msg.Body)
	if err != nil {
		return 0, errors.Wrap(err, "msg.buffer: error while retrieving size of reader")
	}
	if n > maxBodyLen {
		return uint16(n), errors.New("reader body length larger than maximum")
	}
	msg.Body = buf
	return uint16(n), nil
}

// bodyLen returns the size of the body if it can find it, it returns 0, false when it isn't defined
func (msg *Message) bodyLen() (uint16, bool) {
	// If we have no header or message is nil, we return 0, false
	if msg == nil || msg.header == nil {
		return 0, false
	}

	// If the len is indicated in the header, use it
	if bLen := msg.header.bodyLen(); bLen != 0 {
		return uint16(bLen), true
	}

	// Establish the interfaces
	type lener interface {
		Len() int
	}
	type byteser interface {
		Bytes() int
	}

	// Switch over the interfaces
	switch r := msg.Body.(type) {
	case lener:
		return uint16(r.Len()), true
	case byteser:
		return uint16(r.Bytes()), true
	}

	// If we didn't find anything, return 0, false
	return 0, false
}

// WriteTo writes a Message to the given io.Writer
// If m.Body has a Len or Bytes method, then it is used, if it doesn't the request is copied to a buffer before writing it over.
// This consumes the Message Body.
func (msg *Message) WriteTo(w io.Writer) (int64, error) {
	// Check if message is valid
	if msg.header == nil {
		return 0, errors.New("WriteTo: cannot write message as header is nil")
	}

	// If the header has a set content-length, simply use it. If it doesn't, extract it from reader.
	bodyLen, ok := msg.bodyLen()
	if !ok {
		n, err := msg.buffer()
		if err != nil {
			return 0, err
		}
		bodyLen = n
	}
	msg.header.setBodyLen(bodyLen)

	// Now write the header to the writer
	n1, err1 := msg.header.WriteTo(w)
	if err1 != nil && err1 != io.EOF {
		return n1, errors.Wrap(err1, "Read: error while reading header")
	}

	// And read from body
	n2, err2 := io.CopyN(w, msg.Body, int64(bodyLen))
	if err2 != nil && err2 != io.EOF {
		return n1 + n2, errors.Wrap(err2, "Read: error while reading body")
	}

	return n1 + n2, nil
}

// ReadFrom creates a m.Message from an io.Reader.
func (msg *Message) ReadFrom(r io.Reader) (int64, error) {
	// First we decode the header
	h := &header{}
	n1, err := h.ReadFrom(r)
	if err == io.EOF {
		return n1, err
	} else if err != nil {
		return n1, err
	}
	msg.header = h

	// Now, given the header-indicated size we create a buffer of that size
	bodyLen := h.bodyLen()
	content := make([]byte, bodyLen)
	n2, err := r.Read(content)
	if err != nil {
		return n1 + int64(n2), err
	} else if n2 != bodyLen {
		return n1 + int64(n2), errors.Errorf("Read: reader read less than the expected body length (%d): ILLEGAL", bodyLen)
	}

	// And we create a bufio.Reader from it
	body := bytes.NewReader(content)
	msg.Body = body

	return n1 + int64(n2), nil
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
	// TODO: Embed it in a reader checking for ASCII-only text
	return NewMessage(Operator, r)
}

// NewOperatorMessageString returns a message of Operator type built from the given string
func NewOperatorMessageString(txt string) (*Message, error) {
	r := strings.NewReader(txt)
	msg, err := NewMessage(Operator, r)
	if err != nil {
		return msg, err
	}
	msg.header.setBodyLen(uint16(len(txt)))
	return msg, nil
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
func newSystemMessage(ss *systemSig) (*Message, error) {
	return NewMessage(system, bytes.NewReader(ss[:]))
}
