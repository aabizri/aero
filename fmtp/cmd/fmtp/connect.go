package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aabizri/aero/fmtp"
	"github.com/urfave/cli"
)

var (
	connectCommand = cli.Command{
		Name:   "connect",
		Usage:  "Connect to a FMTP client",
		Action: connectAction,
		Flags:  connectFlags,
	}
	connectFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Value: "127.0.0.1:9050",
			Usage: "address to connect to",
		},
		cli.StringFlag{
			Name:  "rem",
			Usage: "ID of the remote endpoint",
		},
	}
)

func connectAction(c *cli.Context) error {
	conn, err := client.Connect(context.Background(), c.String("addr"), fmtp.ID(c.String("rem")))
	if err != nil {
		return err
	}
	_ = conn
	fmt.Printf("Connection successful with %s (addr: %s)", c.String("rem"), c.String("addr"))
	time.Sleep(10 * time.Minute)
	return nil
}
