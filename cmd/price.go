package cmd

import (
	"github.com/urfave/cli/v2"
)

var PriceCommand = &cli.Command{
	Name:        "price",
	Usage:       "Get AWS pricing",
	Aliases:     []string{"p"},
	Subcommands: servicesCommand,
}

var servicesCommand = []*cli.Command{
	ec2Command,
	rdsCommand,
	elasticacheCommand,
}
