package fmtp

import (
	"context"
	"errors"
	"log"
	"net"
	"time"
)

var (
	ErrServerClosed = errors.New("server closed")
)

type Handler interface {
	Handle(m *Message)
}

type handlerFunc struct {
	f func(*Message)
}

func (hf *handlerFunc) Handle(m *Message) {
	hf.f(m)
}

func HandlerFunc(f func(*Message)) Handler {
	return &handlerFunc{f}
}

type Server struct {
	// TCP address to listen on
	Addr string

	// Handler to invoke
	Handler Handler

	// Timeouts
	Ti time.Duration
	Ts time.Duration
	Tr time.Duration

	// Done
	done chan struct{}

	// Client
	c *Client
}

// NewServer creates a server for use
func (c *Client) NewServer(addr string, handler Handler) *Server {
	return &Server{c: c, Addr: addr, Handler: handler}
}

// ListenAndServe listens to an IP Address, and handles functions
func (srv *Server) ListenAndServe() error {
	// Resolve the address
	laddr, err := net.ResolveTCPAddr("tcp", srv.Addr)
	if err != nil {
		return err
	}

	// First, bind to the given address
	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	// Now launch serve on it
	return srv.Serve(l)
}

// logf logs server errors
func (srv *Server) logf(format string, params ...interface{}) {
	log.Printf(format, params...)
}

// Serve serves incoming connections on a net.Listener
func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	var tempDelay time.Duration
	//baseCtx := context.Background()
	for {
		rw, e := l.Accept() //rw, e
		if e != nil {
			select {
			case <-srv.done:
				return ErrServerClosed
			default:
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				srv.logf("http: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0

		// NOW DO SOMETHING WITH THE CONN
		//srv.registerTCPConn(rw)
		_ = rw
	}
}

// Shutdown stops the server gracefully
func (srv *Server) Shutdown(ctx context.Context) error {
	// Closes all open listeners (stopping new connections from forming)
	// Send the Shutdown to all idle Connections
	// Waits for every Connection, idle or associated, to close which can be forever
	return nil
}

// Close stops the server
func (srv *Server) Close() error {
	// Closes all active listeners and all connections, associated or not
	return nil
}
