// Connection establishment overview:
// 	- First a TCP transport is established with the remote FMTP entity
// 	- The responder starts a local timer Ti.
// 	- The initator sends an identification message, and starts a local timer Ti.
// 	- The responder validates the received Identification message, replies by sending an identification message back to the initiator
//	and resetting Ti
// 	- The initiator validates the received identification message, sends an Identification ACCEPT to the responder, and stops Ti
// 	- The responder receives the ACCEPT and stops Ti
// 	- Both endpoints confirm that the connection is established

package fmtp

import (
	"bytes"
	"context"
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrConnectionDeadlineExceeded is returned when the connection deadline (Ti) is exceeded
	ErrConnectionDeadlineExceeded = errors.New("connection deadline exceeded")

	// ErrConnectionRejectedByRemote is returned when the connection has been rejected by the remote party
	ErrConnectionRejectedByRemote = errors.New("connection rejected by remote party")

	// ErrConnectionRejectedByLocal is returned when the connection has been rejected by the local party
	ErrConnectionRejectedByLocal = errors.New("connection rejected for invalid credentials")
)

// Connection holds the connection with an endpoint
type Connection struct {
	// the underlying tcp conn
	tcp *net.TCPConn

	// remote endpoint's ID
	remID ID

	// association if there is one
	// check for termination of association as well
	ass *Association

	// which client does this belong to ?
	client *Client
}

// Close closes the connection
func (conn *Connection) Close() error {
	return conn.tcp.Close()
}

// Shutdown closes the connection gracefully, by also closing the Association if there is one
func (conn *Connection) Shutdown(ctx *context.Context) error {
	return nil
}

func establishTCPConn(ctx context.Context, dialer *net.Dialer, address string) (*net.TCPConn, error) {
	// Establish TCP connection
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	// Assert it as a TCP conn
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, errors.New("establishTCPConn: net.Conn isn't net.TCPConn")
	}

	// Set the connection to the appropriate options
	err = tcpConn.SetKeepAlive(false) //TMP: should be true
	if err != nil {
		return nil, errors.Wrap(err, "establishTCPConn: error while setting keep-alive")
	}

	return tcpConn, nil
}

// Connect creates an FMTP Connection
// If the given context expires before the connection is complete, an error is returned.
// But once successfully established, the context has no effect.
func (c *Client) Connect(ctx context.Context, address string, id ID) (*Connection, error) {
	// Create the TCP connection
	tcpConn, err := establishTCPConn(ctx, c.dialer, address)
	if err != nil {
		return nil, errors.Wrap(err, "Connect: error while establishing TCP connection")
	}
	c.log.Printf("Connection [%s] (%s): successfully established connection", address, id)

	// Create a timer
	// This is a local timer, so we defer its stop
	ti := time.NewTimer(c.tiDuration)
	defer ti.Stop()

	// Create an identification message
	msg, err := newIDRequestMessage(c.id, id)
	if err != nil {
		return nil, errors.Wrap(err, "Connect: error while creating identification message")
	}

	// Send the identification message
	doneChan := make(chan struct{})
	errChan := make(chan error)
	go func() {
		defer func() {
			close(doneChan)
			close(errChan)
		}()

		n, err := msg.WriteTo(tcpConn)
		if err != nil {
			errChan <- errors.Wrapf(err, "Connect: error writing message to tcp connection, wrote %d bytes", n)
		}
		doneChan <- struct{}{}
	}()
	select {
	case <-doneChan:
		break
	case err := <-errChan:
		return nil, err

	case <-ti.C:
		return nil, ErrConnectionDeadlineExceeded
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
	ti.Reset(c.tiDuration)
	c.log.Printf("Connection [%s] (%s): successfully sent identification request", address, id)

	// Receive the reply
	resp := &Message{}
	doneChan = make(chan struct{})
	errChan = make(chan error)
	go func() {
		defer func() {
			close(doneChan)
			close(errChan)
		}()

		_, err := resp.ReadFrom(tcpConn)
		if err != nil {
			errChan <- errors.Wrapf(err, "Connect: error while receiving")
		}
		doneChan <- struct{}{}
	}()
	select {
	case <-doneChan:
		break
	case err := <-errChan:
		return nil, err

	case <-ti.C:
		return nil, ErrConnectionDeadlineExceeded
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
	c.log.Printf("Connection [%s] (%s): successfully received identification request", address, id)

	// Unmarshal the replyMain
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, msg.Body)
	if err != nil {
		return nil, err
	}
	idr := &idRequest{}
	idr.UnmarshalBinary(buf.Bytes())

	// Validate it and create the reply to be sent
	ok := idr.validateID(c.id, id)
	var reply *Message
	switch ok {
	case true:
		// Send an ACCEPT messageMain
		reply, err = newIDResponseMessage(true)
	case false:
		// Send a REJECT message
		reply, err = newIDResponseMessage(false)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Connect: error while creating ID response messages")
	}
	c.log.Printf("Connection [%s] (%s): identification response validated, result is %s", address, id, ok)

	// Send it
	doneChan = make(chan struct{})
	errChan = make(chan error)
	go func() {
		defer func() {
			close(doneChan)
			close(errChan)
		}()

		n, err := reply.WriteTo(tcpConn)
		if err != nil {
			errChan <- errors.Wrapf(err, "Connect: error writing reply to tcp connection, wrote %d bytes", n)
		}
		doneChan <- struct{}{}
	}()
	select {
	case <-doneChan:
		break
	case err := <-errChan:
		return nil, err

	case <-ti.C:
		return nil, ErrConnectionDeadlineExceeded
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
	ti.Stop()
	c.log.Printf("Connection [%s] (%s): successfully sent identification response", address, id)

	// If that was a reject, return an error
	if !ok {
		return nil, ErrConnectionRejectedByLocal
	}

	// Done !
	conn := &Connection{
		tcp:    tcpConn,
		remID:  id,
		client: c,
	}

	c.log.Printf("Connection [%s] (%s) successfully established", address, id)
	return conn, nil
}

// send sends a message over a connection with the given context
func (conn *Connection) send(ctx context.Context, msg *Message) error {
	doneChan := make(chan struct{})
	errChan := make(chan error)
	go func() {
		defer func() {
			close(doneChan)
			close(errChan)
		}()

		n, err := msg.WriteTo(conn.tcp)
		if err != nil {
			errChan <- errors.Wrapf(err, "send: error writing message to tcp connection, wrote %d bytes", n)
		}
		doneChan <- struct{}{}
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

// receive receives a message from the connection
func (conn *Connection) receive(ctx context.Context) (*Message, error) {
	return nil, nil
}
