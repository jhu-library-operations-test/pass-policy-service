package main

import (
	"fmt"

	"github.com/urfave/cli"
)

type serveOpts struct {
	publicBaseURI  string
	privateBaseURI string
	username       string
	passwd         string
	port           int
}

func serve() cli.Command {
	opts := serveOpts{}

	return cli.Command{
		Name:  "serve",
		Usage: "Serve the PASS policy service over http",
		Description: `
			An optional configuration file may be provided as an argument
		`,
		ArgsUsage: "[ file ]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "external, e",
				Usage:       "External (public) PASS baseuri",
				EnvVar:      "PASS_EXTERNAL_FEDORA_BASEURL",
				Destination: &opts.publicBaseURI,
			},
			cli.StringFlag{
				Name:        "internal, i",
				Usage:       "Internal (private) PASS baseuri",
				EnvVar:      "PASS_FEDORA_BASEURL",
				Destination: &opts.privateBaseURI,
			},
			cli.StringFlag{
				Name:        "username, u",
				Usage:       "Username for basic auth to Fedora",
				EnvVar:      "PASS_FEDORA_USER",
				Destination: &opts.username,
			},
			cli.StringFlag{
				Name:        "password, p",
				Usage:       "Password for basic auth to Fedora",
				EnvVar:      "PASS_FEDORA_PASSWORD",
				Destination: &opts.passwd,
			},
			cli.IntFlag{
				Name:        "port",
				Usage:       "Port for the policy service http endpoint",
				EnvVar:      "POLICY_SERVICE_PORT",
				Destination: &opts.port,
			},
		},
		Action: func(c *cli.Context) error {
			return serveAction(opts, c.Args())
		},
	}
}

func serveAction(opts serveOpts, args []string) error {
	return fmt.Errorf("not implemented")
}
