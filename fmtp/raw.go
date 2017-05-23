package fmtp

import (
	"context"
	"io"
	"net"
)

// send is the function that sends a message over a io.Writer
func send(ctx context.Context, w io.Writer, msg *Message) error {
	// Create the local context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Send
	var (
		doneChan = make(chan struct{})
		errChan  = make(chan error)
	)
	go func() {
		defer func() {
			close(doneChan)
			close(errChan)
		}()

		for i := 0; i < 3; i++ {
			_, err := msg.WriteTo(w)

			// Check if this is a temporary error, in such case, retry
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else if err != nil {
				errChan <- err
			} else {
				doneChan <- struct{}{}
			}

			// Return
			return
		}
	}()
	select {
	case <-doneChan:
		break
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// receive is the function that receives a message from a io.Reader
func receive(ctx context.Context, r io.Reader) (*Message, error) {
	// Create the local context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create the message to unmarshal to
	resp := &Message{}

	// Receive
	var (
		doneChan = make(chan struct{})
		errChan  = make(chan error)
	)
	go func() {
		defer func() {
			close(doneChan)
			close(errChan)
		}()

		// Unmarshal from the connection
		_, err := resp.ReadFrom(r)
		switch err {
		case io.EOF: // do nothing
		case nil:
			doneChan <- struct{}{}
		default:
			errChan <- err
		}
	}()
	select {
	case <-doneChan:
		break
	case err := <-errChan:
		return nil, err

	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}

	return resp, nil
}

// disconnect is the actual action taken by an agent when disconnecting
func (conn *Conn) disconnect(ctx context.Context) error {
	return nil
}
