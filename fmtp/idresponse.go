package fmtp

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
)

const (
	accept     = "ACCEPT"
	reject     = "REJECT"
	keywordLen = 6
)

// An idResponse (Identification Message where the messages are being sent to respond to a validation request)
type idResponse struct {
	OK bool
}

// newidResponse returns an ID response
func newidResponse(accept bool) *idResponse {
	return &idResponse{accept}
}

// Reader returns an io.Reader of a binary version of idResponse
func (idr *idResponse) Reader() (io.Reader, error) {
	buf := &bytes.Buffer{}
	b, _ := idr.MarshalBinary()
	n, err := buf.Write(b)
	if err != nil {
		return buf, errors.Wrap(err, "error writing buffer")
	} else if n != len(b) {
		return buf, errors.New("couldn't write all of the bytes")
	}

	return buf, nil
}

// MarshalBinary marshals the idResponse
func (idr *idResponse) MarshalBinary() ([]byte, error) {
	switch idr.OK {
	case true:
		return []byte(accept), nil
	case false:
		return []byte(reject), nil
	}
	panic("unreachable code")
}

// UnmarshalBinary unmarshals the idResponse
func (idr *idResponse) UnmarshalBinary(in []byte) error {
	if len(in) != keywordLen {
		return errors.New("message not to correct size")
	}

	var val idResponse
	switch string(in) {
	case accept:
		val.OK = true
	case reject:
		val.OK = false
	default:
		return errors.New("message is neither ACCEPT nor REJECT")
	}

	idr = &val
	return nil
}
