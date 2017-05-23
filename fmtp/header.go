package fmtp

import (
	"fmt"
	"io"

	"encoding/binary"

	"github.com/pkg/errors"
)

const (
	version2  = 2
	reserved2 = 0

	headerLen = 5
	// maxLength that can be indicated
	maxLength = 32761
	// maxBodyLen is the maximum body len
	maxBodyLen = maxLength - headerLen
)

// header is a FMTP's message Header field
type header struct {
	// version indicated the header version.
	version uint8

	// reserved field is an internal value
	reserved uint8

	// length indicates the combined length in bytes of the Header and Body
	length uint16

	// typ indicates the message type that is being transferred
	typ uint8
}

func (h *header) Check() error {
	if h == nil {
		return nil
	}
	if h.length < headerLen {
		return errors.New("header.Check(): error, indicated length cannot be smaller than nominal header length")
	}
	return nil
}

// String prints the header
func (h *header) String() string {
	return fmt.Sprintf("Version:\t%d\nReserved:\t%d\nLength:\t%d bytes\n\tTyp:\t\t%d\n", h.version, h.reserved, h.length, h.typ)
}

// MarshalBinary marshals a header into binary form
func (h *header) MarshalBinary() ([]byte, error) {
	// Check
	err := h.Check()
	if err != nil {
		return nil, err
	}

	// Get the length in binary
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, h.length)

	// Now create the byte slice
	out := []byte{
		byte(h.version),
		byte(h.reserved),
		byte(lenBuf[0]),
		byte(lenBuf[1]),
		byte(h.typ),
	}
	return out, nil
}

func (h *header) UnmarshalBinary(b []byte) error {
	if len(b) != headerLen {
		return errors.Errorf("UnmarshalBinary: expected %d bytes, got %d", headerLen, len(b))
	}

	// Extract length
	length := binary.BigEndian.Uint16(b[2:4])
	if length > maxLength {
		return errors.New("UnmarshalBinary: indicated length larger than max length")
	} else if length < headerLen {
		return errors.New("UnmarshalBinary: indicated length smaller than nominal header length")
	}

	// Assign
	h.version = b[0]
	h.reserved = b[1]
	h.length = uint16(length)
	h.typ = b[4]

	return nil
}

func (h *header) ReadFrom(r io.Reader) (int64, error) {
	// First read up to headerLen
	b := make([]byte, headerLen)
	n, err := r.Read(b)
	if err == io.EOF {
		return int64(n), io.EOF
	} else if err != nil {
		return int64(n), errors.Wrap(err, "ReadFrom: error while reading from io.Reader")
	} else if n != headerLen {
		return int64(n), errors.New("ReadFrom: couldn't read until header length")
	}

	// Now unmarshal
	err = h.UnmarshalBinary(b)
	if err != nil {
		return int64(n), errors.Wrap(err, "ReadFrom: error while unmarshalling header in binary form")
	}

	return int64(n), nil
}

// WriteTo writes a header to a given io.Writer in binary form
func (h *header) WriteTo(w io.Writer) (int64, error) {
	b, err := h.MarshalBinary()
	if err != nil {
		return 0, errors.Wrap(err, "WriteTo: error while marshalling header to binary")
	}
	n, err := w.Write(b)
	if err != nil {
		return int64(n), errors.Wrap(err, "WriteTo: error while writing to given io.Writer")
	}
	return int64(n), nil
}

// newHeader creates a new header in version 2.0
func newHeader(typ uint8) *header {
	return &header{
		version:  version2,
		reserved: reserved2,
		typ:      typ,
	}
}

// setLength sets the header length
func (h *header) setBodyLen(bodyLen uint16) {
	h.length = headerLen + bodyLen
}

// bodyLen returns the body length
// if no body is here
func (h *header) bodyLen() int {
	if h.length == 0 {
		return 0
	}
	return int(h.length) - headerLen
}
