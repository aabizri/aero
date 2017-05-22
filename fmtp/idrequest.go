package fmtp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const (
	hyphen          = "-"
	hyphenLen       = len("-")
	minIDRequestLen = minIDLen*2 + hyphenLen + keywordLen
	maxIDRequestLen = maxIDLen*2 + hyphenLen + keywordLen
)

// An idRequest (Identification Message where the messages are being sent for validation)
type idRequest struct {
	// Identification Value of the sending system
	Sender ID
	// Identification Value of the receiving system
	Receiver ID
}

func (idr *idRequest) String() string {
	return fmt.Sprintf("sender is %s, receiver is %s", idr.Sender, idr.Receiver)
}

// newidRequest creates an idRequest
func newidRequest(sender ID, receiver ID) (*idRequest, error) {
	// Check the parameters
	errS, errR := sender.Check(), receiver.Check()
	switch {
	case errS != nil:
		return nil, errors.Wrap(errS, "NewidRequest: invalid sender ID")
	case errR != nil:
		return nil, errors.Wrap(errR, "NewidRequest: invalid receiver ID")
	}

	// Return the MSG as a reader
	return &idRequest{
		Sender:   sender,
		Receiver: receiver,
	}, nil
}

// validateID reports whether the ids match
func (idr *idRequest) validateID(sender, receiver ID) bool {
	return idr.Sender == sender && idr.Receiver == receiver
}

func (idr *idRequest) Len() int {
	return len(idr.Sender) + 1 + len(idr.Receiver)
}

// MarshalBinary marshals the idRequest
func (idr *idRequest) MarshalBinary() ([]byte, error) {
	// We know the length has a known minimum and maximum length, so we set the length to the minimum, and the capacity to the maximum.
	output := make([]byte, 0, maxIDRequestLen)

	output = append(output, []byte(idr.Sender)...)
	output = append(output, byte('-'))
	output = append(output, []byte(idr.Receiver)...)

	return output, nil
}

// UnmarshalBinary unmarshals a idRequest
func (idr *idRequest) UnmarshalBinary(b []byte) error {
	if len(b) > maxIDRequestLen {
		return errors.Errorf("id request length exceeds maximum (%d > %d)", len(b), maxIDRequestLen)
	}

	// Let's split the byte slice into two
	parts := bytes.Split(b, []byte(hyphen))
	if len(parts) != 2 {
		return errors.New("invalid id request: less or more than 1 hyphen in")
	}

	// Assign
	idr.Sender = ID(parts[0])
	idr.Receiver = ID(parts[1])
	return nil
}

// WriteTo implements io.WriterTo
func (idr *idRequest) WriteTo(w io.Writer) (int64, error) {
	bin, err := idr.MarshalBinary()
	if err != nil {
		return 0, errors.Wrap(err, "WriteTo: error while marshalling to binary")
	}
	n, err := w.Write(bin)
	if err != nil {
		return int64(n), err
	}
	if n != len(bin) {
		return int64(n), errors.Errorf("WriteTo: binary form has len %d, wrote %d: mismatch", len(bin), n)
	}
	return int64(n), nil
}

// Reader returns an io.Reader
func (idr *idRequest) Reader() (io.Reader, error) {
	buf := &bytes.Buffer{}
	_, err := idr.WriteTo(buf)
	return buf, err
}
