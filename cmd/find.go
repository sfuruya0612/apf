package cmd

import (
	"fmt"

	"github.com/sfuruya0612/apf/internal/mongo"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
)

var FindCommand = &cli.Command{
	Name:    "get",
	Usage:   "Get AWS pricing",
	Aliases: []string{"g"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "service",
			Aliases: []string{"serv"},
			EnvVars: []string{"SERVICE"},
			Value:   "ec2",
			Usage:   "Specify a valid AWS service name",
		},
		&cli.StringFlag{
			Name:    "spec",
			Aliases: []string{"s"},
			EnvVars: []string{"SPEC"},
			Value:   "t3.micro",
			Usage:   "Specify a valid AWS service spec",
		},
	},
	Action: func(ctx *cli.Context) error {
		return get(ctx.String("mongo-uri"), ctx.String("service"), ctx.String("spec"))
	},
}

func get(mongoUri, service, spec string) error {
	conn, err := mongo.Connect(mongoUri)
	if err != nil {
		return fmt.Errorf("Failed to connect to MongoDB: %w", err)
	}

	coll := mongo.Collection(conn, service)

	filter := bson.M{"product.attributes.instancetype": spec}

	results, err := mongo.Find(coll, filter, nil)
	if err != nil {
		return fmt.Errorf("Failed to find: %w", err)
	}

	fmt.Println(results)

	if err := mongo.Disconnect(conn); err != nil {
		return fmt.Errorf("Failed to disconnect to MongoDB: %w", err)
	}
	return nil
}
