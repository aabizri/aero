package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aabizri/aero/fmtp"
	"github.com/urfave/cli"
)

var (
	associateCommand = cli.Command{
		Name:   "associate",
		Usage:  "Associate with a remote",
		Flags:  associateFlags,
		Action: associateAction,
	}
	associateFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Usage: "address to associate to",
			Value: "127.0.0.1:9050",
		},
		cli.StringFlag{
			Name:  "rem",
			Usage: "id of the remote server",
		},
	}
)

func associateAction(c *cli.Context) error {
	ctx := context.Background()
	conn, err := client.Connect(ctx, c.String("addr"), fmtp.ID(c.String("rem")))
	if err != nil {
		return err
	}
	fmt.Printf("Connection successful with %s (addr: %s), now associating...\n", c.String("rem"), c.String("addr"))

	err = conn.Associate(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Association successful !")

	time.Sleep(10 * time.Minute)
	return nil
}
