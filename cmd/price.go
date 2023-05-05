package cmd

import (
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
)

var PriceCommand = &cli.Command{
	Name:    "price",
	Usage:   "Get AWS pricing",
	Aliases: []string{"p"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "instance-type",
			Aliases: []string{"i"},
			Usage:   "Specify a valid instance type",
		},
		&cli.StringFlag{
			Name:    "vcpu",
			Aliases: []string{"cpu"},
			Usage:   "Specify a valid vCPU",
		},
		&cli.StringFlag{
			Name:    "memory",
			Aliases: []string{"mem"},
			Usage:   "Specify a valid memory",
		},
	},
	Subcommands: servicesCommand,
}

var servicesCommand = []*cli.Command{
	ec2Command,
	rdsCommand,
	elasticacheCommand,
}

func appendCondition(filter bson.M, instanceType, vcpu, memory string) bson.M {
	if instanceType != "" {
		filter["product.attributes.instancetype"] = instanceType
	}

	if vcpu != "" {
		filter["product.attributes.vcpu"] = vcpu
	}

	if memory != "" {
		filter["product.attributes.memory"] = memory + " GiB"
	}

	return filter
}
