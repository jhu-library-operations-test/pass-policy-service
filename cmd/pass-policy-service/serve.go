package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/oa-pass/pass-policy-service/web"
	"github.com/pkg/errors"
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

	if len(args) != 1 {
		return fmt.Errorf("expecting exactly one argument: the rules doc file")
	}

	var credentials *web.Credentials
	if opts.username != "" {
		credentials = &web.Credentials{
			Username: opts.username,
			Password: opts.passwd,
		}
	}

	rules, err := ioutil.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("error reading %s: %s", args[0], err.Error())
	}

	policyService, err := web.NewPolicyService(rules, &web.InternalPassClient{
		Requester:       &http.Client{},
		ExternalBaseURI: opts.publicBaseURI,
		InternalBaseURI: opts.privateBaseURI,
		Credentials:     credentials,
	})
	if err != nil {
		return errors.Wrapf(err, "could not initialize policy service")
	}

	policyService.Replace = web.BaseURIs{
		Public:  opts.publicBaseURI,
		Private: opts.privateBaseURI,
	}

	http.HandleFunc("/policies", policyService.RequestPolicies)
	http.HandleFunc("/repositories", policyService.RequestRepositories)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opts.port))
	if err != nil {
		return err
	}

	opts.port = listener.Addr().(*net.TCPAddr).Port
	log.Printf("Listening on port %d", opts.port)

	return http.Serve(listener, nil)
}
