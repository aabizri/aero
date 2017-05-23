package fmtp

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

var (
	// ErrAssociationTimeoutExceeded happens when the association reception timeout (tr) is exceeded
	// It is returned when associating. Once associated, such error will shutdown the association.
	ErrAssociationTimeoutExceeded = errors.New("association reception timeout exceeded")
)

// initAssociate establishes an FMTP association, without locking !
func (conn *Conn) initAssociate(ctx context.Context) error {
	// Create a STARTUP request
	msg, err := newSystemMessage(startup)
	if err != nil {
		return errors.Wrap(err, "Associate: error while creating system message")
	}

	// Send it
	err = conn.send(ctx, msg)
	if err != nil {
		return err
	}
	fmt.Println("SENT STARTUP")

	// Wait for a STARTUP response
	trCtx, cancel := context.WithTimeout(context.Background(), conn.trDuration)
	defer cancel()
	reply, err := conn.receive(trCtx)
	if trCtx.Err() != nil { // If the cancel comes from tiCtx, we do not return a "context canceled" but the correct error
		return ErrAssociationTimeoutExceeded
	} else if err != nil {
		return err
	}
	if reply.header.typ != system {
		return errors.New("unexpected reply type")
	}
	fmt.Println("RECEIVED SYSTEM MSG")

	// Unmarshal it
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, reply.Body)
	if err != nil {
		return err
	}
	ss := &systemSig{}
	err = ss.UnmarshalBinary(buf.Bytes())
	if err != nil {
		return err
	}

	// Check if its startup
	if !ss.equals(startup) {
		return errors.New("system message not startup")
	}
	fmt.Println("CONFIRMED AS STARTUP")

	return nil
}

// recvAssociate establishes an association requested by the peer
func (conn *Conn) recvAssociate(ctx context.Context) error {
	// TODO: Check for acceptance by the user

	// Create a STARTUP request
	msg, err := newSystemMessage(startup)
	if err != nil {
		return errors.Wrap(err, "Associate: error while creating system message")
	}

	// Send it
	err = conn.send(ctx, msg)
	if err != nil {
		return err
	}

	// OK
	return nil
}

// deassociate is the actual action taken by an agent when deassociating
func (conn *Conn) deassociate(ctx context.Context) error {
	return nil
}
