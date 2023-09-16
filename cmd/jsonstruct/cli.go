package main

import "github.com/urfave/cli/v2"

func httpCommand() *cli.Command {
	return &cli.Command{
		Name:        "http",
		Action:      httpListener,
		Usage:       "run a web app to generate structs in the browser",
		Description: "This will run a web app to let you generate new structs in your browser, listening on localhost:8080 by default.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "the `HOST` to listen on",
				Value: "127.0.0.1",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "the `PORT` to listen on",
				Value: 8080,
			},
		},
	}
}
