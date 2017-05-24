package fmtp_test

import (
	"io"
	"testing"
)

type nopCloser struct {
	io.ReadWriter
}

func (nc nopCloser) Close() error {
	return nil
}

func TestConnect_Offline(t *testing.T) { /*
		// Create the buffer
		buf := nopCloser{&bytes.Buffer{}}

		// Write
		conn := defaultClient.NewConn()
		err := conn.SetUnderlying(buf)*/
}
