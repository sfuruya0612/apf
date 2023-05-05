package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/sfuruya0612/apf/internal/mongo"
	"github.com/sfuruya0612/apf/internal/utils"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var elasticacheCommand = &cli.Command{
	Name:    "elasticache",
	Aliases: []string{"ec"},
	Usage:   "Get Elasticache pricing",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "engine",
			Aliases: []string{"e"},
			Value:   "Redis",
			Usage:   "Specify a valid cache engine (e.g. Redis, Memcached)",
		},
	},
	Action: func(ctx *cli.Context) error {
		return getElasticachePrice(ctx.String("mongo-uri"), ctx.String("instance-type"), ctx.String("vcpu"), ctx.String("memory"), ctx.String("engine"))
	},
}

func getElasticachePrice(mongoUri, instanceType, vcpu, memory, engine string) error {
	conn, err := mongo.Connect(mongoUri)
	if err != nil {
		return fmt.Errorf("Failed to connect to MongoDB: %w", err)
	}

	coll := mongo.Collection(conn, "elasticache")

	filter := bson.M{"product.attributes.osengine": engine}

	filter = appendCondition(filter, instanceType, vcpu, memory)

	results, err := mongo.Find(coll, filter, nil)
	if err != nil {
		return fmt.Errorf("Failed to find: %w", err)
	}

	printElasticache(results)

	if err := mongo.Disconnect(conn); err != nil {
		return fmt.Errorf("Failed to disconnect to MongoDB: %w", err)
	}
	return nil
}

func printElasticache(results []bson.M) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)

	if _, err := fmt.Fprintln(w, strings.Join(getElasticacheHeader(), "\t")); err != nil {
		return fmt.Errorf("Failed to print header: %w", err)
	}

	for _, result := range results {
		if _, err := fmt.Fprintln(w, formatElasticache(result)); err != nil {
			return fmt.Errorf("Failed to print result: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("Failed to flush: %w", err)
	}

	return nil
}

func getElasticacheHeader() []string {
	return []string{
		"Service",
		"Region",
		"OS/Engine",
		"InstanceType",
		"vCPU",
		"Memory",
		"OnDemandPrice(USD/hour)",
		"OnDemandPrice(USD/month)",
	}
}

func formatElasticache(result primitive.M) string {
	fields := []string{
		result["servicecode"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["regioncode"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["osengine"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["instancetype"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["vcpu"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["memory"].(string),
		result["ondemandpriceperusd"].(string),
		utils.ConvertHourlyToMonthly(result["ondemandpriceperusd"].(string)),
	}

	return strings.Join(fields, "\t")
}
