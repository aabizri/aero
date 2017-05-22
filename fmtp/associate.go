package fmtp

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// Associate establishes an FMTP association
func (conn *Connection) Associate(ctx context.Context) error {
	// We lock the association field of a connection in Write Mode, blocking all who whould wish to use it
	conn.assMu.Lock()
	defer conn.assMu.Unlock()

	// If we are already associated, we return as such
	if conn.Associated() {
		return nil
	}

	// We launch the non-locking associate
	return conn.associate(ctx)
}

// associate establishes an FMTP association, without locking !
func (conn *Connection) associate(ctx context.Context) error {
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

	// Create the tr timer
	tr := time.NewTimer(conn.client.trDuration)

	// Wait for a STARTUP response
	// As tr is the timer here, and we don't want to mix ctx expiration with tr timer expiration, we do it in a goroutine
	var reply *Message
	doneChan := make(chan struct{})
	errChan := make(chan error)
	lctx, lcancel := context.WithCancel(context.Background())
	defer lcancel()
	go func() {
		defer func() {
			close(doneChan)
			close(errChan)
		}()
		reply, err = conn.receive(lctx)
		if err != nil {
			errChan <- err
		}
		doneChan <- struct{}{}
	}()
	select {
	case <-doneChan:
		break
	case err := <-errChan:
		return err

	case <-tr.C:
		return ErrassociationTimeoutExceeded
	case <-ctx.Done():
		lcancel()
		return ctx.Err()
	}

	// Parse it
	// TODO

	// Create the ts timer
	a := &association{
		tsDuration: conn.client.tsDuration,
		trDuration: conn.client.trDuration,
		conn:       conn,
	}
	a.tr = time.AfterFunc(a.trDuration, func() {
		a.Shutdown(context.Background())
	})
	a.ts = time.AfterFunc(a.tsDuration, a.heartbeater)
	conn.ass = a

	return nil
}
