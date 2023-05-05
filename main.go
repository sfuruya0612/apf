package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sfuruya0612/apf/cmd"
	"github.com/urfave/cli/v2"
)

var (
	commit string
)

var Commands = []*cli.Command{
	cmd.FetchCommand,
	cmd.PriceCommand,
}

func main() {
	app := cli.NewApp()

	app.Name = "apf"
	app.Usage = "CLI tool to get AWS pricing information"
	app.Version = fmt.Sprintf("%s (Build by: %s)", commit, runtime.Version())

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "mongo-uri",
			Aliases: []string{"m"},
			EnvVars: []string{"MONGODB_URI"},
			Value:   "mongodb://localhost:27017",
			Usage:   "Specify a valid MongoDB URI",
		},
	}

	app.Commands = Commands

	if err := app.Run(os.Args); err != nil {
		code := 1
		if c, ok := err.(cli.ExitCoder); ok {
			code = c.ExitCode()
		}
		fmt.Printf("\x1b[31mERROR: %v\x1b[0m", err.Error())
		os.Exit(code)
	}
}
