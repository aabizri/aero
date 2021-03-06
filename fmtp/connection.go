package fmtp

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// ErrConnectionDeadlineExceeded is returned when the connection deadline (Ti) is exceeded
	ErrConnectionDeadlineExceeded = errors.New("connection deadline exceeded")

	// ErrConnectionRejectedByRemote is returned when the connection has been rejected by the remote party
	ErrConnectionRejectedByRemote = errors.New("connection rejected by remote party")

	// ErrConnectionRejectedByLocal is returned when the connection has been rejected by the local party
	ErrConnectionRejectedByLocal = errors.New("connection rejected for invalid credentials")
)

// Conn holds the connection with an endpoint
type Conn struct {
	// remote endpoint's ID for connection initalisation
	// when receiving a connection, acceptRemote function is used, which then sets remID
	remote ID
	local  ID

	// acceptRemote is called when receiving a connection, a positive return means the ID has been accepted
	acceptRemote func(ID) bool

	// the underlying tcp conn, or any io.RWC
	tcp io.ReadWriteCloser

	// orders is how an order is given to the agent
	orders chan order

	// done closes the agent directly
	done chan struct{}

	// ti is the maximum period of time in which data must be received during an FMTP connection attempt in order for it to be successful
	tiDuration time.Duration

	// ts is the maximum period of time in which data must be transmitted in order to maintain an FMTP association
	tsDuration time.Duration

	// tr is the maximum period of time in which data is to be received over an FMTP association
	trDuration time.Duration

	// handler is the user's handler for OPERATOR and OPERATIONAL messages
	handler Handler

	// shutdownHandler notifies the user that a shutdown has been initiated
	shutdownHandler func()

	// tx and rx are how much has been outputted/inputted
	tx uint64
	rx uint64

	// which client does this belong to ?
	client *Client
}

// SetTimers sets the connection timers
func (conn *Conn) SetTimers(ti, tr, ts time.Duration) {
	conn.tiDuration = ti
	conn.trDuration = tr
	conn.tsDuration = ts
}

// SetUnderlying sets the underlying connection.
// The protocol requires TCP connection. However, for debugging, tunneling or other usecases, it can be beneficial to set a custom one.
func (conn *Conn) SetUnderlying(rwc io.ReadWriteCloser) error {
	if rwc == nil {
		return errors.New("SetUnderlying: given io.ReadWriteCloser is nil, can't set")
	}
	conn.tcp = rwc
	return nil
}

// SetAcceptRemote sets the function that accepts remote IDs for incoming connections
func (conn *Conn) SetAcceptRemote(f func(ID) bool) error {
	if f == nil {
		return errors.New("SetAcceptRemote: given function is nil, can't set")
	}
	conn.acceptRemote = f
	return nil
}

// NewConn creates a new connection
func (c *Client) NewConn() *Conn {
	// Establish the defaults
	conn := &Conn{
		local:      c.id,
		orders:     make(chan order),
		done:       make(chan struct{}),
		tiDuration: c.tiDuration,
		trDuration: c.trDuration,
		tsDuration: c.tsDuration,
		client:     c,
	}

	return conn
}

// Init initialises a connection
func (conn *Conn) Init(ctx context.Context, addr string, remote ID) error {
	// Debug
	logger := conn.client.logger.WithFields(logrus.Fields{
		"addr":      addr,
		"remote ID": remote,
	})
	logger.Debug("Conn.Init called")

	// Set the remote indicated here as the conn's remote
	conn.remote = remote

	// If there is no underlying connection set, create a TCP connection
	if conn.tcp == nil {
		logger.Debug("no underlying connection set, establishing a TCP connection now...")
		// Create the TCP connection
		tcpConn, err := establishTCPConn(ctx, conn.client.dialer, addr)
		if err != nil {
			return errors.Wrap(err, "Connect: error while establishing TCP connection")
		}
		conn.tcp = tcpConn
	}

	// Send an ID Request
	err := conn.sendIDRequestMessage(ctx, conn.local, remote)
	if err != nil {
		return err
	}

	// Create a new context for us to be able to cancel execution, it will act as the ti timer.
	tiCtx, cancel := context.WithTimeout(ctx, conn.tiDuration)
	defer cancel()

	// Receive an ID Request, using the tiCtx
	idr, err := conn.recvIDRequestMessage(tiCtx)
	if tiCtx.Err() != nil { // If the cancel comes from tiCtx, we do not return a "context canceled" but the correct error
		return ErrConnectionDeadlineExceeded
	} else if err != nil {
		return err
	}

	// Validate it and send the reply, using the tiCtx
	ok := idr.validateID(remote, conn.local)
	err = conn.sendIDResponseMessage(tiCtx, ok)
	if tiCtx.Err() != nil { // If the cancel comes from tiCtx, we do not return a "context canceled" but the correct error
		return ErrConnectionDeadlineExceeded
	} else if err != nil {
		return err
	}

	// If that was a reject, return an error
	if !ok {
		return ErrConnectionRejectedByLocal
	}

	// Launch the agent
	go conn.agent()

	// Register the connection client-side
	err = conn.client.registerConn(conn)
	if err != nil {
		return err
	}

	// Finished
	return nil
}

// Close closes the association & connection without any grace
func (conn *Conn) Close() error {
	conn.done <- struct{}{}
	return nil
}

// Disconnect disconnects a connection, gracefully
func (conn *Conn) Disconnect(ctx context.Context) error {
	return conn.order(ctx, disconnectCmd)
}

// Deassociate de-associates gracefully
func (conn *Conn) Deassociate(ctx context.Context) error {
	return conn.order(ctx, deassociateCmd)
}

// Connect initiates an FMTP Connection. It is a wrapper around (*Client).NewConn and NewConn.Init
// If the given context expires before the connection is complete, an error is returned.
// But once successfully established, the context has no effect.
func (c *Client) Connect(ctx context.Context, address string, id ID) (*Conn, error) {
	conn := c.NewConn()
	err := conn.Init(ctx, address, id)
	return conn, err
}

// Associate upgrades an FMTP Connection to an association
// If the given context expires before the connection is complete, an error is returned.
// But once successfully established, the context has no effect.
func (conn *Conn) Associate(ctx context.Context) error {
	return conn.order(ctx, associateCmd)
}

// Send sends a message over a connection, making the agent associate it if needed.
func (conn *Conn) Send(ctx context.Context, msg *Message) error {
	done := make(chan error)
	go func() {
		conn.orders <- order{
			command: sendCmd,
			ctx:     ctx,
			done:    done,
			msg:     msg,
		}
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RemoteAddr returns the remote address behind a connection, if there is one
func (conn *Conn) RemoteAddr() net.Addr {
	type remoteAddrer interface {
		RemoteAddr() net.Addr
	}
	if ra, ok := conn.tcp.(remoteAddrer); ok {
		return ra.RemoteAddr()
	}
	return nil
}

// RemoteID returns the ID of the connection's remote party, empty ID if not currently set
func (conn *Conn) RemoteID() ID {
	return conn.remote
}

// send sends a message over a connection, adding the size of the message to the tx value
func (conn *Conn) send(ctx context.Context, msg *Message) error {
	n, err := send(ctx, conn.tcp, msg)
	atomic.AddUint64(&conn.tx, uint64(n))
	return err
}

// receive receives a message from the connection, adding the size of the message to the rx value
func (conn *Conn) receive(ctx context.Context) (*Message, error) {
	msg, err := receive(ctx, conn.tcp)
	if bl, ok := msg.bodyLen(); ok {
		atomic.AddUint64(&conn.rx, uint64(bl))
	}
	return msg, err
}
