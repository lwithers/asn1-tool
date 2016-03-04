package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/lwithers/asn1-tool/html"
	"github.com/lwithers/asn1-tool/util"
)

func main() {
	app := cli.NewApp()
	app.Name = "asn1-tool"
	app.Usage = "investigate ASN.1 DER encoded data"
	app.Commands = []cli.Command{
		{
			Name:   "html",
			Usage:  "display structure in HTML",
			Action: CmdWrapper(html.Html),
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

// CmdWrapper takes a cli.Command.Action-like function which returns an error,
// and builds an actual cli.Command.Action function which reports any error to
// stderr. It provides simple, consistent error handling.
func CmdWrapper(cmd func(c *cli.Context) error) func(c *cli.Context) {
	return func(c *cli.Context) {
		if err := cmd(c); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %v\n",
				c.App.Name, c.Command.Name, err)
			if _, ok := err.(*util.UsageError); ok {
				cli.ShowCommandHelp(c, c.Command.Name)
			}
			os.Exit(1)
		}
	}
}
