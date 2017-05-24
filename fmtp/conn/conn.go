// Package conn allows you to use fmtp as a net.Conn.
package conn

import "github.com/aabizri/aero/fmtp"

// The Conn is the exported type
type Conn struct {
	fmtp *fmtp.Conn
}
