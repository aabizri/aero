package fmtp

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrassociationTimeoutExceeded happens when the association reception timeout (tr) is exceeded
	// It is returned when associating. Once associated, such error will shutdown the association.
	ErrassociationTimeoutExceeded = errors.New("association reception timeout exceeded")
)

// association stores the details of an association
type association struct {
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

	// conn is the underlying connection
	conn *Connection
}

// handle is what we use to serve the message to the user
// 	- It resets the tr timer
// 	- If the message is a HEARTBEAT, then it stops there
// 	- If it is a SHUTDOWN it launches a.Close
// 	- Else, it gives them to a user handler
func (a *association) handle(msg *Message) {
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

// handleErr dispatches an error in the handling to the user
func (a *association) handleErr(err error) {}

// serve is a function launched in its own goroutine that serves incoming messages
func (a *association) serve() {
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

	// Event loop, checking for arrival
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
func (a *association) Shutdown(ctx context.Context) error {
	// Lock the association
	a.conn.assMu.Lock()
	defer a.conn.assMu.Unlock()

	// Let's stop both timers
	a.ts.Stop()
	a.tr.Stop()

	// Now close the watchdog
	a.done <- struct{}{}

	// Create SHUTDOWN message

	// Send it over

	// Select on them all and ctx.Done

	return nil
}

// heartbeater is the function called by the ts ticker
func (a *association) heartbeater() {
	// SEND HEARTBEAT MESSAGE
	a.ts.Reset(a.tsDuration)
}
