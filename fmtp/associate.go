package fmtp

import (
	"bytes"
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// ErrAssociationTimeoutExceeded happens when the association reception timeout (tr) is exceeded
	// It is returned when associating. Once associated, such error will shutdown the association.
	ErrAssociationTimeoutExceeded = errors.New("association reception timeout exceeded")
)

// initAssociate establishes an FMTP association, without locking !
func (conn *Conn) initAssociate(ctx context.Context) error {
	// Debuf
	logger := conn.client.logger.WithFields(logrus.Fields{
		"addr":      conn.RemoteAddr().String(),
		"remote ID": conn.RemoteID(),
	})
	logger.Debug("initAssociate called")

	logger.Debug("creating STARTUP request")
	// Create a STARTUP request
	msg, err := newSystemMessage(startup)
	if err != nil {
		return errors.Wrap(err, "Associate: error while creating system message")
	}

	logger.Debug("sending STARTUP request")
	// Send it
	err = conn.send(ctx, msg)
	if err != nil {
		return err
	}
	logger.Debug("send successful, waiting for response")

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
	logger.Debug("response retrieved")

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

	logger.Debug("connection initiated")
	return nil
}

// recvAssociate establishes an association requested by the peer
func (conn *Conn) recvAssociate(ctx context.Context) error {
	conn.client.logger.Debug("recvAssociate called")
	conn.client.logger.Debug("sending STARTUP")
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
	conn.client.logger.Debug("STARTUP sent")

	// OK
	return nil
}

// deassociate is the actual action taken by an agent when deassociating
func (conn *Conn) deassociate(ctx context.Context) error {
	return nil
}
