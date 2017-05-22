package fmtp

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
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

	// Callback to notify we received a TCP connection
	NotifyTCP func(remoteAddr net.Addr)

	// Callback to notify that a connection was succesfull
	NotifyConn func(remoteAddr net.Addr, remoteID ID)

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
func (srv *Server) Serve(l *net.TCPListener) error {
	defer l.Close()
	var tempDelay time.Duration
	//baseCtx := context.Background()
	for {
		// We accept the next connection
		rw, e := l.AcceptTCP() //rw, e

		// Check for errors
		if e != nil {
			select {
			case <-srv.done:
				return nil
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

		// We notify, if there is a notify function, that a TCP connection has been made
		if srv.NotifyTCP != nil {
			go srv.NotifyTCP(rw.RemoteAddr())
		}

		// We have a new TCP conn, so we register it
		go srv.registerTCPConn(rw)
	}
}

// registerTCPConn registers a new TCP connection
func (srv *Server) registerTCPConn(tcp *net.TCPConn) {
	conn, err := srv.c.recvConnect(context.Background(), tcp, srv.c.id)
	if err != nil {
		fmt.Printf("Error while negotiating incoming connection: %s", err)
		tcp.Write([]byte("ERROR: ILLEGAL\n"))
		tcp.Close()
	}
	if srv.NotifyConn != nil {
		go srv.NotifyConn(conn.tcp.RemoteAddr(), conn.remID)
	}
	_ = conn
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
