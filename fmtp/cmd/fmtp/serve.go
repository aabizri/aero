package main

import (
	"fmt"
	"net"
	"os"

	"github.com/aabizri/aero/fmtp"
	"github.com/urfave/cli"
)

// serveCommand
var (
	serveCommand = cli.Command{
		Name:   "serve",
		Usage:  "Launch a server",
		Action: serveAction,
		Flags:  serveFlags,
	}

	serveFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Value: "127.0.0.1:9050",
			Usage: "address to listen to",
		},
	}
)

// serveAction launches a new server
func serveAction(c *cli.Context) error {
	addr := c.String("addr")
	handler := fmtp.HandlerFunc(
		func(msg *fmtp.Message) {
			msg.WriteTo(os.Stdout)
		})
	srv := client.NewServer(addr, handler)
	srv.NotifyTCP = func(addr net.Addr) {
		fmt.Printf("Server> received new TCP connection from %s\n", addr)
	}
	srv.NotifyConn = func(addr net.Addr, rem fmtp.ID) {
		fmt.Printf("Server> connection established with %s (%s)", rem, addr)
	}
	return srv.ListenAndServe()
}
