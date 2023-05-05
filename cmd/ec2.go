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

var ec2Command = &cli.Command{
	Name:  "ec2",
	Usage: "Get EC2 pricing",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "os",
			Aliases: []string{"o"},
			Value:   "Linux",
			Usage:   "Specify a valid OS (e.g. Linux, RHEL, SUSE, Windows, ...)",
		},
		&cli.StringFlag{
			Name:    "tenancy",
			Aliases: []string{"t"},
			Value:   "Shared",
			Usage:   "Specify a valid tenancy (e.g. Shared, Dedicated, Host, Reserved, NA)",
		},
		&cli.StringFlag{
			Name:    "capacitystatus",
			Aliases: []string{"c"},
			Value:   "Used",
			Usage:   "Specify a valid capacitystatus (e.g. Used, UnusedCapacityReservation, AllocatedCapacityReservation)",
		},
		&cli.StringFlag{
			Name:    "preinstalled-sw",
			Aliases: []string{"p"},
			Value:   "NA",
			Usage:   "Specify a valid preInstalled sw (e.g. NA, SQL Web, SQL Std, ...)",
		},
	},
	Action: func(ctx *cli.Context) error {
		return getEc2Price(ctx.String("mongo-uri"), ctx.String("instance-type"), ctx.String("vcpu"), ctx.String("memory"), ctx.String("os"), ctx.String("tenancy"), ctx.String("capacitystatus"), ctx.String("preinstalled-sw"))
	},
}

func getEc2Price(mongoUri, instanceType, vcpu, memory, os, tenancy, capacitystatus, preInstalledSw string) error {
	conn, err := mongo.Connect(mongoUri)
	if err != nil {
		return fmt.Errorf("Failed to connect to MongoDB: %w", err)
	}

	coll := mongo.Collection(conn, "ec2")

	filter := bson.M{"product.attributes.osengine": os, "product.attributes.tenancy": tenancy, "product.attributes.capacitystatus": capacitystatus, "product.attributes.preinstalledsw": preInstalledSw}

	filter = appendCondition(filter, instanceType, vcpu, memory)

	results, err := mongo.Find(coll, filter, nil)
	if err != nil {
		return fmt.Errorf("Failed to find: %w", err)
	}

	printEc2(results)

	if err := mongo.Disconnect(conn); err != nil {
		return fmt.Errorf("Failed to disconnect to MongoDB: %w", err)
	}
	return nil
}

func printEc2(results []bson.M) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)

	if _, err := fmt.Fprintln(w, strings.Join(getEc2Header(), "\t")); err != nil {
		return fmt.Errorf("Failed to print header: %w", err)
	}

	for _, result := range results {
		if _, err := fmt.Fprintln(w, formatEc2(result)); err != nil {
			return fmt.Errorf("Failed to print result: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("Failed to flush: %w", err)
	}

	return nil
}

func getEc2Header() []string {
	return []string{
		"Service",
		"Region",
		"OS/Engine",
		"InstanceType",
		"vCPU",
		"Memory",
		"PhysicalProcessor",
		"ClockSpeed(GHz)",
		"Tenancy",
		"CapacityStatus",
		"PreInstalledSw",
		"ProcessorArchitecture",
		"OnDemandPrice(USD/hour)",
		"OnDemandPrice(USD/month)",
	}
}

func formatEc2(result primitive.M) string {
	fields := []string{
		result["servicecode"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["regioncode"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["osengine"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["instancetype"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["vcpu"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["memory"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["physicalprocessor"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["clockspeed"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["tenancy"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["capacitystatus"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["preinstalledsw"].(string),
		result["product"].(bson.M)["attributes"].(bson.M)["processorarchitecture"].(string),
		result["ondemandpriceperusd"].(string),
		utils.ConvertHourlyToMonthly(result["ondemandpriceperusd"].(string)),
	}

	return strings.Join(fields, "\t")
}
