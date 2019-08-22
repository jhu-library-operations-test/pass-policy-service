package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/oa-pass/pass-policy-service/rule"
	"github.com/urfave/cli"
)

func validate() cli.Command {

	return cli.Command{
		Name:  "validate",
		Usage: "Validate policy rules files",
		Description: `
			Given a list of policy rules files, validate will attempt to parse 
			the documents and validate it with respect to the schema used by this
			policy service.

			Note, the documents will be validated against schemas supported by this 
			application regardless of any schema declarations in the file. 
		`,
		ArgsUsage: "files",
		Action: func(c *cli.Context) error {
			return validateAction(c.Args())
		},
	}
}

func validateAction(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("validate requires at least one schema")
	}

	var lastErr error

	for _, instance := range args {
		content, err := ioutil.ReadFile(instance)
		if err != nil {
			lastErr = err
			continue
		}

		_, err = rule.Validate(content)
		if err == nil {
			log.Printf("Validation OK: %s", instance)
		} else {
			errtxt := strings.ReplaceAll(fmt.Sprintf("%v", err), "\n", "\n  ")
			log.Printf("Validation failed: %s:\n  %v", instance, errtxt)
			lastErr = err
		}
	}

	return lastErr
}
