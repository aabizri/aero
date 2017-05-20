package fmtp

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrAssociationTimeoutExceeded happens when the association reception timeout (tr) is exceeded
	// It is returned when associating. Once associated, such error will shutdown the association.
	ErrAssociationTimeoutExceeded = errors.New("association reception timeout exceeded")
)

// Association stores the details of an association
type Association struct {
	// The underlying connection
	conn *Connection

	// done channel closes the watchdog
	done chan struct{}

	// ts is the maximum period of time in which data must be transmitted in order to maintain the FMTP connection
	tsMu       sync.Mutex
	tsDuration time.Duration
	ts         *time.Timer

	// tr is the maximum period of time in which data is to be received over an FMTP association
	trMu       sync.Mutex
	trDuration time.Duration
	tr         *time.Timer

	// handler is the user's handler for OPERATOR and OPERATIONAL messages
	handler func(m *Message)

	// shutdownHandler notifies the user that a shutdown has been initiated
	shutdownHandler func()

	// terminated means the association has been terminated
	terminatedMu sync.Mutex
	terminated   bool
}

// handle is what we use to serve the message to the user
// 	- It resets the tr timer
// 	- If the message is a HEARTBEAT, then it stops there
// 	- If it is a SHUTDOWN it launches a.Close
// 	- Else, it gives them to a user handler
func (a *Association) handle(msg *Message) {
	// First, reset the timer
	a.tr.Reset(a.trDuration)

	// Now, if it is a system message, we check the embedded message
	switch msg.header.typ {
	case system:
		// parse the body
		// if it is heartbeat we return here
		// if it is shutdown we call shutdownHandler, followed by a.terminate
	case Operator, Operational:
		if a.handler != nil {
			a.handler(msg)
		}
	}
}

func (a *Association) handleErr(err error) {}

// watchdog is a function launched in its own goroutine that serves incoming messages and resets
func (a *Association) watchdog() {
	// Create a watchdog context
	ctx, cancel := context.WithCancel(context.Background())

	// Launch a listener goroutine
	msgChan := make(chan *Message, 3)
	errChan := make(chan error)
	go func(ctx context.Context, mc <-chan *Message, ec <-chan error) {
		// Listen to incoming requests.
		// For each new request, launch a goroutine that unmarshals them and then outputs them to mc
		// Don't listen to context here, just pass it on to the spawned goroutines
	}(ctx, msgChan, errChan)

	// Check for arrival
	for {
		select {
		// If we received a message, we handle it
		case msg := <-msgChan:
			a.handle(msg)

		// If we received an error, we evaluate it
		case err := <-errChan:
			a.handleErr(err)

		// In case we get killed, we stop
		case <-a.done:
			cancel()
			return
		}
	}
}

// Shutdown closes an association by sending the shutdown message
func (a *Association) Shutdown(ctx context.Context) error {
	// First, let's stop both timers
	a.ts.Stop()
	a.tr.Stop()

	// Now close the watchdog
	a.done <- struct{}{}

	// Create SHUTDOWN message

	// Send it over

	// Select on them all and ctx.Done

	return nil
}

// ShutdownRecurse closes an association and its underlying connection at the same time
func (a *Association) ShutdownRecurse(ctx context.Context) error {
	return nil
}

// heartbeater is the function called by the ts ticker
func (a *Association) heartbeater() {
	// SEND HEARTBEAT MESSAGE
	a.ts.Reset(a.tsDuration)
}

// Associate establishes an FMTP Association
func (conn *Connection) Associate(ctx context.Context) (*Association, error) {
	// Send a STARTUP request
	// Create an identification message
	msg, err := newSystemMessage(startup)
	if err != nil {
		return nil, errors.Wrap(err, "Connect: error while creating identification message")
	}

	// Send it
	err = conn.send(ctx, msg)
	if err != nil {
		return nil, err
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
		return nil, err

	case <-tr.C:
		return nil, ErrAssociationTimeoutExceeded
	case <-ctx.Done():
		lcancel()
		return nil, ctx.Err()
	}

	// Parse it
	// TODO

	// Create the ts timer
	a := &Association{
		tsDuration: conn.client.tsDuration,
		trDuration: conn.client.trDuration,
		conn:       conn,
	}
	a.tr = time.AfterFunc(a.trDuration, func() {
		a.Shutdown(context.Background())
	})
	a.ts = time.AfterFunc(a.tsDuration, a.heartbeater)
	return a, nil

}

// Send sends a message
//
// The context cancels the send if it gets cancelled while sending
func (a *Association) Send(ctx context.Context, msg *Message) error {
	// Send a message
	n, err := msg.WriteTo(a.conn.tcp)
	if err != nil {
		return errors.Wrapf(err, "Send: error writing message to tcp connection, wrote %d bytes", n)
	}

	// Reset ts timer
	a.ts.Reset(a.tsDuration)
	return nil
}

// SendOperational sends an operational message
// If r has a Len or Bytes method, it is more efficient, so try to use strings.Reader, buf.Reader, buf.Buffer or others with those methods.
func (a *Association) SendOperational(ctx context.Context, r io.Reader) error {
	// Create the operational message
	msg, err := NewOperationalMessage(r)
	if err != nil {
		return err
	}

	// Copy it to the included tcp connection
	return a.Send(ctx, msg)
}

// SendOperator sends an operator message
// If r has a Len or Bytes method, it is more efficient, so try to use strings.Reader, buf.Reader, buf.Buffer or others with those methods.
func (a *Association) SendOperator(ctx context.Context, r io.Reader) error {
	// Create the operational message
	msg, err := NewOperatorMessage(r)
	if err != nil {
		return err
	}

	// Copy it to the included tcp connection
	return a.Send(ctx, msg)
}
