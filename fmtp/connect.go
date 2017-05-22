// connect.go manages the process of connecting
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

	"github.com/pkg/errors"
)

// sendIDRequestMessage sends an IDRequestMessage
func (conn *Connection) sendIDRequestMessage(ctx context.Context, local ID, remote ID) error {
	// Create an identification message
	msg, err := newIDRequestMessage(local, remote)
	if err != nil {
		return err
	}

	// Send the identification message
	return conn.send(ctx, msg)
}

// recvIDRequestMessage receives an IDRequestMessage and extracts it from the message
func (conn *Connection) recvIDRequestMessage(ctx context.Context) (*idRequest, error) {
	// Receive the reply
	msg, err := conn.receive(ctx)
	if err != nil {
		return nil, err
	}

	// If it isn't an ID message, it's an error
	if msg.header != nil && msg.header.typ != identification {
		return nil, errors.New("received message isn't of correct typ")
	}

	// Unmarshal the reply body
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, msg.Body)
	if err != nil {
		return nil, err
	}
	idr := &idRequest{}
	idr.UnmarshalBinary(buf.Bytes())

	// Return it
	return idr, nil
}

// sendIDResponseMessage sends an Identification Response message.
func (conn *Connection) sendIDResponseMessage(ctx context.Context, ok bool) error {
	// Create an identification message
	msg, err := newIDResponseMessage(ok)
	if err != nil {
		return errors.Wrap(err, "Connect: error while creating identification message")
	}

	// Send the identification message
	return conn.send(ctx, msg)
}

// recvIDResponseMessage receives an an Identification Response message and unmarshals it
func (conn *Connection) recvIDResponseMessage(ctx context.Context) (*idResponse, error) {
	// Receive the reply
	msg, err := conn.receive(ctx)

	// Unmarshal the reply
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, msg.Body)
	if err != nil {
		return nil, err
	}
	idresp := &idResponse{}
	err = idresp.UnmarshalBinary(buf.Bytes())
	if err != nil {
		return nil, err
	}

	// Return it
	return idresp, nil
}

// initConnect initiates a connection with a remote party
func (c *Client) initConnect(ctx context.Context, tcp *net.TCPConn, local ID, remote ID) (*Connection, error) {
	// First, create the connection
	conn := &Connection{
		tcp:    tcp,
		remID:  remote,
		client: c,
	}

	// Send an ID Request
	err := conn.sendIDRequestMessage(ctx, local, remote)
	if err != nil {
		return nil, err
	}

	// Create a new context for us to be able to cancel execution, it will act as the ti timer.
	tiCtx, cancel := context.WithTimeout(ctx, c.tiDuration)
	defer cancel()

	// Receive an ID Request, using the tiCtx
	idr, err := conn.recvIDRequestMessage(tiCtx)
	if tiCtx.Err() != nil { // If the cancel comes from tiCtx, we do not return a "context canceled" but the correct error
		return nil, ErrConnectionDeadlineExceeded
	} else if err != nil {
		return nil, err
	}

	// Validate it and send the reply, using the tiCtx
	ok := idr.validateID(remote, local)
	err = conn.sendIDResponseMessage(tiCtx, ok)
	if tiCtx.Err() != nil { // If the cancel comes from tiCtx, we do not return a "context canceled" but the correct error
		return nil, ErrConnectionDeadlineExceeded
	} else if err != nil {
		return nil, err
	}

	// If that was a reject, return an error
	if !ok {
		return nil, ErrConnectionRejectedByLocal
	}

	// If it's OK, then we're done !
	return conn, nil
}

// recvConnection receives a connection request from an outside party
func (c *Client) recvConnect(ctx context.Context, tcp *net.TCPConn, local ID) (*Connection, error) {
	// First, create the connection
	conn := &Connection{
		tcp:    tcp,
		client: c,
	}

	// We create a local context following the ti timer
	tiCtx, cancel := context.WithTimeout(ctx, c.tiDuration)
	defer cancel()

	// Receive an ID Request, using the tiCtx
	c.log("recvConnect: receiving id request message...")
	idr, err := conn.recvIDRequestMessage(tiCtx)
	if tiCtx.Err() != nil { // If the cancel comes from tiCtx, we do not return a "context canceled" but the correct error
		return nil, ErrConnectionDeadlineExceeded
	} else if err != nil {
		return nil, err
	}
	c.log("recvConnect: received ID request message")

	// We validate it and send our request, we should have a function in Client that checks if its acceptable
	// Plus we should send a REJECT message if that isn't right !
	/*ok := idr.validateID(remote, local)
	if !ok {
		return nil, ErrConnectionRejectedByLocal
	}*/

	// We note the remote ID in the connection
	conn.remID = idr.Sender

	// We send an ID request message using the normal context
	err = conn.sendIDRequestMessage(ctx, local, idr.Sender)
	if err != nil {
		return nil, err
	}
	c.log("recvConnect: sent ID request message")

	// We reset our tiCtx
	tiCtx, cancel = context.WithTimeout(ctx, c.tiDuration)
	defer cancel()

	// We await a positive response
	idresp, err := conn.recvIDResponseMessage(tiCtx)
	if tiCtx.Err() != nil { // If the cancel comes from tiCtx, we do not return a "context canceled" but the correct error
		return nil, ErrConnectionDeadlineExceeded
	} else if err != nil {
		return nil, err
	}

	// If the response was negative, we signal it
	if !idresp.OK {
		return nil, ErrConnectionRejectedByRemote
	}

	return conn, nil
}
