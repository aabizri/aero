package fmtp

import (
	"context"
	"io"
	"net"

	"github.com/pkg/errors"
)

// send is the function that sends a message over a io.Writer
func send(ctx context.Context, w io.Writer, msg *Message) (int, error) {
	// Create the local context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Send
	var (
		ret = make(chan error)
		n   int64
	)
	defer close(ret)
	go func() {
		var err error
		for i := 0; i < 3; i++ {
			n, err = msg.WriteTo(w)

			// Check if this is a temporary error, in such case, retry
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				ret <- err
			}

			// Return
			return
		}
		ret <- errors.Wrap(err, "send: cannot send after 3 retry")
	}()
	select {
	case err := <-ret:
		return int(n), err
	case <-ctx.Done():
		return int(n), ctx.Err()
	}
}

// receive is the function that receives a message from a io.Reader
func receive(ctx context.Context, r io.Reader) (*Message, error) {
	// Create the message to unmarshal to
	resp := &Message{}

	// Receive
	ret := make(chan error)
	defer close(ret)
	go func() {

		// Unmarshal from the connection
		_, err := resp.ReadFrom(r)
		ret <- err
	}()
	select {
	case err := <-ret:
		return resp, err

	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// disconnect is the actual action taken by an agent when disconnecting
func (conn *Conn) disconnect(ctx context.Context) error {
	return nil
}
