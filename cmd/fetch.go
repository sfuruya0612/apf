package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sfuruya0612/apf/internal/aws"
	"github.com/sfuruya0612/apf/internal/mongo"
	"github.com/urfave/cli/v2"
)

var (
	serviceCodes = []string{"AmazonEC2", "AmazonRDS", "AmazonElastiCache"}
	// skuOfferTermCode := fmt.Sprintf("%s.%s", sku, "JRTCKXETXF")
	// skuOfferTermCodeRateCode := fmt.Sprintf("%s.%s.%s", sku, "JRTCKXETXF", "6YS6EN2CT7")
)

var FetchCommand = &cli.Command{
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
	},
	Action: func(ctx *cli.Context) error {
		return fetch(ctx.String("profile"), ctx.String("region"), ctx.String("mongo-uri"))
	},
}

func fetch(profile, region, mongoUri string) error {
	cfg, err := aws.Config(profile, region)
	if err != nil {
		return fmt.Errorf("Fetch: %w", err)
	}

	// It takes about 20 minutes to get and insert EC2 price
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	errCh := make(chan error, len(serviceCodes))
	wg := sync.WaitGroup{}
	wg.Add(len(serviceCodes))
	sem := make(chan struct{}, 10)

	for _, sc := range serviceCodes {
		go func(sc string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			products, err := aws.GetProducts(cfg, sc)
			if err != nil {
				errCh <- fmt.Errorf("Failed to fetch %s products: %w", sc, err)
				return
			}

			conn, err := mongo.Connect(mongoUri)
			if err != nil {
				errCh <- fmt.Errorf("Failed to connect to MongoDB: %w", err)
				return
			}

			coll := mongo.Collection(conn, getCollectionName(sc))

			log.Printf("Drop %s collection\n", sc)

			if err := mongo.DropCollection(coll, nil); err != nil {
				errCh <- fmt.Errorf("Failed to remove %s collection: %w", sc, err)
				return
			}

			log.Printf("Inserting %d %s products into MongoDB\n", len(products), sc)

			// TODO: Bulk insert
			var insertErr error
			for _, product := range products {
				if _, err := coll.InsertOne(ctx, product); err != nil {
					insertErr = fmt.Errorf("Failed to insert %s product %s: %w", sc, *product, err)
					break
				}
			}

			if insertErr != nil {
				errCh <- insertErr
				return
			}

			if err := mongo.Disconnect(conn); err != nil {
				errCh <- fmt.Errorf("Failed to disconnect to MongoDB: %w", err)
				return
			}

			log.Printf("Inserted %d %s products\n", len(products), sc)
		}(sc)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	log.Println("Completed saving AWS Price List data to MongoDB")

	return nil
}

func getCollectionName(serviceCode string) string {
	switch serviceCode {
	case "AmazonEC2":
		return "ec2"
	case "AmazonRDS":
		return "rds"
	case "AmazonElastiCache":
		return "elasticache"
	default:
		panic(fmt.Sprintf("Unknown service code: %s", serviceCode))
	}
}
