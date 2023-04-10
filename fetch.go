package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	serviceCodes = []string{"AmazonEC2", "AmazonRDS", "AmazonElastiCache"}
)

func Fetch(profile, region, mongoUrl string) error {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return fmt.Errorf("Failed to load AWS config: %v", err)
	}

	clientOptions := options.Client().ApplyURI(mongoUrl)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return fmt.Errorf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	if err := fetchAll(ctx, cfg, client); err != nil {
		return fmt.Errorf("Failed to fetch all products: %v", err)
	}

	log.Println("Completed saving AWS Price List data to MongoDB")

	return nil
}

func fetchAll(ctx context.Context, cfg aws.Config, client *mongo.Client) error {
	errCh := make(chan error, len(serviceCodes))
	wg := sync.WaitGroup{}
	wg.Add(len(serviceCodes))
	sem := make(chan struct{}, 10)

	for _, serviceCode := range serviceCodes {
		go func(serviceCode string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			// Fetch products from AWS Price List API
			products, err := fetchProducts(cfg, serviceCode)
			if err != nil {
				errCh <- fmt.Errorf("Failed to fetch %s products: %v", serviceCode, err)
				return
			}

			log.Printf("Inserting %d %s products into MongoDB\n", len(products), serviceCode)
			collection := client.Database("aws_pricing").Collection(getCollectionName(serviceCode))
			var insertErr error
			for _, product := range products {
				log.Printf("Inserting %v product into MongoDB\n", &product)
				if _, err := collection.InsertOne(ctx, product); err != nil {
					insertErr = fmt.Errorf("Failed to insert %s product: %v", serviceCode, err)
					break
				}
			}
			if insertErr != nil {
				errCh <- insertErr
				return
			}

			log.Printf("Inserted %d %s products\n", len(products), serviceCode)
		}(serviceCode)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchProducts(cfg aws.Config, serviceCode string) ([]*pricing.GetProductsOutput, error) {
	log.Printf("Fetching %s products from AWS Price List API\n", serviceCode)

	var products []*pricing.GetProductsOutput

	client := pricing.NewFromConfig(cfg)

	input := &pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
	}

	paginator := pricing.NewGetProductsPaginator(client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get products: %v", err)
		}

		products = append(products, output)
	}

	return products, nil
}

func parseProduct(price string) (map[string]interface{}, error) {
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(price), &product); err != nil {
		return nil, err
	}
	return product, nil
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
		return ""
	}
}
