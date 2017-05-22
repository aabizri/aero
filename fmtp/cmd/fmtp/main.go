package main

import (
	"fmt"
	"os"

	"github.com/aabizri/aero/fmtp"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "local",
		Value: "localID",
		Usage: "ID for the local client",
	},
}

var commands = []cli.Command{
	cli.Command{
		Name:  "send",
		Usage: "Send an FMTP message",
	},
	connectCommand,
	serveCommand,
	associateCommand,
}

var (
	client *fmtp.Client
)

func main() {
	app := cli.NewApp()
	app.Name = "fmtp"
	app.Usage = "send or listen to fmtp messages, the Flight Message Transfer Protocol"
	app.Flags = flags
	app.Commands = commands
	app.Before = func(c *cli.Context) error {
		lid := fmtp.ID(c.String("local"))
		cl, err := fmtp.NewClient(lid)
		if err != nil {
			return err
		}
		client = cl
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
