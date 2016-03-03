package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "asn1-tool"
	app.Usage = "investigate ASN.1 DER encoded data"
	app.Commands = []cli.Command{
		{
			Name:   "html",
			Usage:  "display structure in HTML",
			Action: CmdWrapper(Html),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "out",
					Usage: "name of output file",
				},
			},
		},
	}
	app.RunAndExitOnError()
}

// UsageError should be returned by a command action if there is a problem with
// the arguments. The error within will be reported to the user, and help for
// the given action will be displayed.
type UsageError struct {
	Problem string
}

func (u *UsageError) Error() string {
	return u.Problem
}

// CmdWrapper takes a cli.Command.Action-like function which returns an error,
// and builds an actual cli.Command.Action function which reports any error to
// stderr. It provides simple, consistent error handling.
func CmdWrapper(cmd func(c *cli.Context) error) func(c *cli.Context) {
	return func(c *cli.Context) {
		if err := cmd(c); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %v\n",
				c.App.Name, c.Command.Name, err)
			if _, ok := err.(*UsageError); ok {
				cli.ShowCommandHelp(c, c.Command.Name)
			}
			os.Exit(1)
		}
	}
}
