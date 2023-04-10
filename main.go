package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"
)

var (
	commit string
)

func main() {
	app := cli.NewApp()

	app.Name = "apf"
	app.Usage = "CLI tool to get AWS pricing information"
	app.Version = fmt.Sprintf("%s (Build by: %s)", commit, runtime.Version())

	// app.Flags = []cli.Flag{
	//         &cli.StringFlag{
	//                 Name:    "service",
	//                 Aliases: []string{"s"},
	//                 EnvVars: []string{"SERVICE"},
	//                 Value:   "AmazonEC2",
	//                 Usage:   "Specify a valid AWS service e.g. AmazonEC2, AmazonRDS",
	//         },
	//         &cli.StringFlag{
	//                 Name:    "region",
	//                 Aliases: []string{"r"},
	//                 EnvVars: []string{"AWS_REGION"},
	//                 Value:   "ap-northeast-1",
	//                 Usage:   "Specify a valid AWS region",
	//         },
	// }

	// app.Before = cmd.Before

	app.Commands = []*cli.Command{
		{
			Name:    "fetch",
			Usage:   "Fetch AWS pricing information",
			Aliases: []string{"f"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "profile",
					Aliases: []string{"p"},
					EnvVars: []string{"AWS_PROFILE"},
					Value:   "default",
					Usage:   "Specify a valid AWS profile",
				},
				&cli.StringFlag{
					Name:    "region",
					Aliases: []string{"r"},
					EnvVars: []string{"AWS_REGION"},
					Value:   "us-east-1",
					Usage:   "Specify a valid AWS region",
				},
				&cli.StringFlag{
					Name:    "mongo-uri",
					Aliases: []string{"m"},
					EnvVars: []string{"MONGODB_URI"},
					Value:   "mongodb://localhost:27017",
					Usage:   "Specify a valid MongoDB URI",
				},
			},
			Action: func(ctx *cli.Context) error {
				return Fetch(ctx.String("profile"), ctx.String("region"), ctx.String("mongo-uri"))
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		code := 1
		if c, ok := err.(cli.ExitCoder); ok {
			code = c.ExitCode()
		}
		fmt.Printf("\x1b[31mERROR: %v\x1b[0m", err.Error())
		os.Exit(code)
	}
}
