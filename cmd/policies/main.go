package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

var fatalf = log.Fatalf

func main() {
	app := cli.NewApp()
	app.Name = "policies"
	app.Usage = "PASS policy service"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		serve(),
		validate(),
	}
	err := app.Run(os.Args)
	if err != nil {
		fatalf("%s", err)
	}
}
