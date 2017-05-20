package fmtp

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	"time"

	"github.com/pkg/errors"
)

const (
	defaultTiDuration = 12 * time.Second
	defaultTsDuration = 12 * time.Second
	defaultTrDuration = 12 * time.Second
)

// Client is what allows you to do FMTP requests.
type Client struct {
	dialer *net.Dialer
	id     ID

	// log is the default logger
	log *log.Logger

	// default timer durations
	tiDuration time.Duration
	tsDuration time.Duration
	trDuration time.Duration

	// currentConns map IDs to ongoing connections
	mu           sync.Mutex
	currentConns map[ID]*Connection
}

type ClientSetter func(c *Client) error

// SetDialer sets a dialer
func SetDialer(dialer *net.Dialer) ClientSetter {
	return func(c *Client) error {
		c.dialer = dialer
		return nil
	}
}

// SetTimers sets the timers
// 	ti is the connection timer, it is only used when establishing connections
// 	ts is the ...
// 	tr is the ...
func SetTimers(ti, ts, tr time.Duration) ClientSetter {
	return func(c *Client) error {
		if ti != 0 {
			c.tiDuration = ti
		}
		if ts != 0 {
			c.tsDuration = ts
		}
		if tr != 0 {
			c.trDuration = tr
		}
		return nil
	}
}

// NewClient creates a new FMTP client
func NewClient(id ID, setters ...ClientSetter) (*Client, error) {
	// Validate the ID
	err := id.Check()
	if err != nil {
		return nil, err
	}

	// Create the default client
	c := &Client{
		id:           id,
		dialer:       &net.Dialer{},
		log:          log.New(os.Stdout, "fmtp-debug> ", 0),
		tiDuration:   defaultTiDuration,
		tsDuration:   defaultTsDuration,
		trDuration:   defaultTrDuration,
		currentConns: map[ID]*Connection{},
	}

	// Now apply the setters
	for _, s := range setters {
		err = s(c)
		if err != nil {
			return c, err
		}
	}

	return c, nil
}

// Dial Connects and Associates with a remote FMTP responder
//
// FMTP dialing has two steps: first connect, then associate.
func (c *Client) Dial(ctx context.Context, address string, id ID) (*Association, error) {
	conn, err := c.Connect(ctx, address, id)
	if err != nil {
		return nil, errors.Wrap(err, "Dial: error while establishing connection")
	}

	ass, err := conn.Associate(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Dial: error while establishing association ")
	}

	return ass, nil
}
