package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/sfuruya0612/apf/internal/utils"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var rdsCommand = &cli.Command{
	Name:  "rds",
	Usage: "Get RDS pricing",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "engine",
			Aliases: []string{"e"},
			Value:   "Aurora MySQL",
			Usage:   "Specify a valid databaes engine (e.g. Aurora MySQL, MySQL, Aurora PostgreSQL, PostgreSQL, MariaDB, Oracle, SQLServer)",
		},
		&cli.StringFlag{
			Name:    "deployment-option",
			Aliases: []string{"d"},
			Value:   "Single-AZ",
			Usage:   "Specify a valid deployment option (e.g. Singe-AZ, Multi-AZ)",
		},
	},
	Action: func(ctx *cli.Context) error {
		return getRdsPrice(ctx)
	},
}

func getRdsPrice(ctx *cli.Context) error {
	filter := bson.M{
		"product.attributes.osengine":         ctx.String("engine"),
		"product.attributes.deploymentoption": ctx.String("deployment-option"),
	}

	results, err := findMongo(
		ctx.String("mongo-uri"),
		"rds",
		ctx.String("instance-type"),
		ctx.String("vcpu"),
		ctx.String("memory"),
		filter,
	)
	if err != nil {
		return fmt.Errorf("Failed to find: %w", err)
	}

	printRds(results)

	return nil
}

func printRds(results []bson.M) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)

	if _, err := fmt.Fprintln(w, strings.Join(getRdsHeader(), "\t")); err != nil {
		return fmt.Errorf("Failed to print header: %w", err)
	}

	for _, result := range results {
		if _, err := fmt.Fprintln(w, formatRds(result)); err != nil {
			return fmt.Errorf("Failed to print result: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("Failed to flush: %w", err)
	}

	return nil
}

func getRdsHeader() []string {
	return []string{
		"Service",
		"Region",
		"OS/Engine",
		"InstanceType",
		"vCPU",
		"Memory",
		"DeploymentOption",
		"Storage",
		"OnDemandPrice(USD/hour)",
		"OnDemandPrice(USD/month)",
	}
}

func formatRds(result primitive.M) string {
	fields := []string{
		result["servicecode"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["regioncode"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["osengine"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["instancetype"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["vcpu"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["memory"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["deploymentoption"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["storage"].(string),
		result["ondemandpriceperusd"].(string),
		utils.ConvertHourlyToMonthly(result["ondemandpriceperusd"].(string)),
	}

	return strings.Join(fields, "\t")
}
