package cmd

import (
	"fmt"

	"github.com/sfuruya0612/apf/internal/mongo"
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

func findMongo(mongoUri, collection, instanceType, vcpu, memory string, filter bson.M) ([]bson.M, error) {
	conn, err := mongo.Connect(mongoUri)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to MongoDB: %w", err)
	}

	coll := mongo.Collection(conn, collection)

	f := appendCondition(filter, instanceType, vcpu, memory)

	results, err := mongo.Find(coll, f, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to find: %w", err)
	}

	if err := mongo.Disconnect(conn); err != nil {
		return nil, fmt.Errorf("Failed to disconnect to MongoDB: %w", err)
	}

	// I'm not sure about returning it with an error
	if len(results) == 0 {
		return nil, fmt.Errorf("No results")
	}

	return results, nil
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
