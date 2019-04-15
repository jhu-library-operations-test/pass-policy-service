package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/oa-pass/pass-policy-service/rule"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func validate() cli.Command {

	return cli.Command{
		Name:  "validate",
		Usage: "Validate a given policy rules file",
		Description: `
			Given a policy rules file, validate will attempt to parse the document
			and validate it with respect to the schema used by this polivy service.

			Note, the document will be validated against schemas supported by this 
			application regardless of any schema declarations in the file. 
		`,
		ArgsUsage: "file",
		Action: func(c *cli.Context) error {
			return validateAction(c.Args())
		},
	}
}

func validateAction(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("validate expects exactly one argument")
	}

	content, err := ioutil.ReadFile(args[0])
	if err != nil {
		return errors.Wrapf(err, "error opening file")
	}

	err = rule.Validate(content)
	if err == nil {
		log.Println("Validation OK")
	}

	return err
}
