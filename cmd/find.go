package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/sfuruya0612/apf/internal/mongo"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	printResults(results)

	if err := mongo.Disconnect(conn); err != nil {
		return fmt.Errorf("Failed to disconnect to MongoDB: %w", err)
	}
	return nil
}

func printResults(results []bson.M) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)

	header := []string{
		"Service",
		"Region",
		"Engine",
		"InstanceType",
		"vCPU",
		"Memory",
		"PricePerUSD",
	}

	if _, err := fmt.Fprintln(w, strings.Join(header, "\t")); err != nil {
		return fmt.Errorf("Failed to print header: %w", err)
	}

	for _, result := range results {
		if _, err := fmt.Fprintln(w, format(result)); err != nil {
			return fmt.Errorf("Failed to print result: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("Failed to flush: %w", err)
	}

	return nil
}

func format(result primitive.M) string {
	fields := []string{
		result["servicecode"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["regioncode"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["engine"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["instancetype"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["vcpu"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["memory"].(string),
		result["priceperusd"].(string),
	}

	return strings.Join(fields, "\t")
}
