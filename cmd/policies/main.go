package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "policies"
	app.Usage = "PASS policy service"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		serve(),
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
