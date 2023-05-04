package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
)

type Pricing struct {
	Product struct {
		ProductFamily string
		Attributes    struct {
			Memory                  string
			Vcpu                    string
			InstanceType            string
			UsageType               string
			LocationType            string
			InstanceFamily          string
			Engine                  string
			RegionCode              string
			Servicecode             string
			CurrentGeneration       string
			NetworkPerformance      string
			Location                string
			Servicename             string
			Operation               string
			EngineCode              string
			InstanceTypeFamily      string
			Storage                 string
			NormalizationSizeFactor string
			DatabaseEdition         string
			PhysicalProcessor       string
			LicenseModel            string
			DeploymentOption        string
			ProcessorArchitecture   string
		}
	}
	ServiceCode string
	PricePerUSD string
}

func FetchPricing(cfg aws.Config, serviceCode string) ([]*Pricing, error) {
	log.Printf("Fetching %s products from AWS Price List API\n", serviceCode)

	client := pricing.NewFromConfig(cfg)

	input := &pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
		Filters: []types.Filter{
			{
				Field: aws.String("regionCode"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String("ap-northeast-1"),
			},
			// Only AWS Region location. (Exclude AWS Outpost)
			{
				Field: aws.String("locationType"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String("AWS Region"),
			},
		},
	}

	paginator := pricing.NewGetProductsPaginator(client, input)

	var p []*Pricing

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("Failed to get products: %v", err)
		}

		p, err = parsePricing(serviceCode, p, output.PriceList)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse products: %v", err)
		}

	}

	return p, nil
}

func parsePricing(serviceCode string, pricings []*Pricing, priceList []string) ([]*Pricing, error) {
	for _, price := range priceList {
		p, err := parseProduct(price)
		if err != nil {
			return nil, err
		}

		// If the product does not have vcpu or memory, skip it.
		if p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["vcpu"] == nil || p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["memory"] == nil {
			continue
		}

		// TODO: Check offerTermCode and rateCode, if not match return error.

		pricing := &Pricing{}
		pricing.ServiceCode = serviceCode
		pricing.Product.ProductFamily = p["product"].(map[string]interface{})["productFamily"].(string)

		// TODO: Looking for a better way to get the price.
		sku := p["product"].(map[string]interface{})["sku"].(string)
		skuOfferTermCode := fmt.Sprintf("%s.%s", sku, "JRTCKXETXF")
		skuOfferTermCodeRateCode := fmt.Sprintf("%s.%s.%s", sku, "JRTCKXETXF", "6YS6EN2CT7")

		pricing.PricePerUSD = p["terms"].(map[string]interface{})["OnDemand"].(map[string]interface{})[skuOfferTermCode].(map[string]interface{})["priceDimensions"].(map[string]interface{})[skuOfferTermCodeRateCode].(map[string]interface{})["pricePerUnit"].(map[string]interface{})["USD"].(string)

		switch serviceCode {
		case "AmazonRDS":
			pricing = rdsPricing(pricing, p)
		case "AmazonElastiCache":
			pricing = elasticachePricing(pricing, p)
		default:
			panic("Unknown service code")
		}

		pricings = append(pricings, pricing)
	}

	return pricings, nil
}

func parseProduct(price string) (map[string]interface{}, error) {
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(price), &product); err != nil {
		return nil, err
	}
	return product, nil
}

// Example:
//
//	"product": {
//	  "productFamily": "Database Instance",
//	  "attributes": {
//	    "engineCode": "5",
//	    "instanceTypeFamily": "M5d",
//	    "memory": "128 GiB",
//	    "vcpu": "32",
//	    "instanceType": "db.m5d.8xlarge",
//	    "usagetype": "APN1-Multi-AZUsage:db.m5d.8xl",
//	    "locationType": "AWS Region",
//	    "storage": "2 x 600 NVMe SSD",
//	    "normalizationSizeFactor": "128",
//	    "instanceFamily": "General purpose",
//	    "databaseEngine": "Oracle",
//	    "databaseEdition": "Enterprise",
//	    "regionCode": "ap-northeast-1",
//	    "servicecode": "AmazonRDS",
//	    "physicalProcessor": "Intel Xeon Platinum 8175",
//	    "licenseModel": "Bring your own license",
//	    "currentGeneration": "Yes",
//	    "networkPerformance": "10 Gbps",
//	    "deploymentOption": "Multi-AZ",
//	    "location": "Asia Pacific (Tokyo)",
//	    "servicename": "Amazon Relational Database Service",
//	    "processorArchitecture": "64-bit",
//	    "operation": "CreateDBInstance:0005"
//	  },
//	  "sku": "22GTWH3M7MMRFZQ9"
//	}
func rdsPricing(pricing *Pricing, p map[string]interface{}) *Pricing {
	// Forcefully accommodating differences between database engines.
	if p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["engineCode"] == nil {
		pricing.Product.Attributes.EngineCode = "unknown"
	} else {
		pricing.Product.Attributes.EngineCode = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["engineCode"].(string)
	}

	if p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["databaseEdition"] == nil {
		pricing.Product.Attributes.DatabaseEdition = "unknown"
	} else {
		pricing.Product.Attributes.DatabaseEdition = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["databaseEdition"].(string)
	}

	if p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["physicalProcessor"] == nil {
		pricing.Product.Attributes.PhysicalProcessor = "unknown"
	} else {
		pricing.Product.Attributes.PhysicalProcessor = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["physicalProcessor"].(string)
	}

	if p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["currentGeneration"] == nil {
		pricing.Product.Attributes.CurrentGeneration = "unknown"
	} else {
		pricing.Product.Attributes.CurrentGeneration = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["currentGeneration"].(string)
	}

	if p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["networkPerformance"] == nil {
		pricing.Product.Attributes.NetworkPerformance = "unknown"
	} else {
		pricing.Product.Attributes.NetworkPerformance = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["networkPerformance"].(string)
	}

	if p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["processorArchitecture"] == nil {
		pricing.Product.Attributes.ProcessorArchitecture = "unknown"
	} else {
		pricing.Product.Attributes.ProcessorArchitecture = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["processorArchitecture"].(string)
	}

	pricing.Product.Attributes.InstanceTypeFamily = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["instanceTypeFamily"].(string)
	pricing.Product.Attributes.Memory = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["memory"].(string)
	pricing.Product.Attributes.Vcpu = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["vcpu"].(string)
	pricing.Product.Attributes.InstanceType = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["instanceType"].(string)
	pricing.Product.Attributes.UsageType = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["usagetype"].(string)
	pricing.Product.Attributes.LocationType = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["locationType"].(string)
	pricing.Product.Attributes.Storage = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["storage"].(string)
	pricing.Product.Attributes.NormalizationSizeFactor = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["normalizationSizeFactor"].(string)
	pricing.Product.Attributes.InstanceFamily = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["instanceFamily"].(string)
	pricing.Product.Attributes.Engine = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["databaseEngine"].(string)
	pricing.Product.Attributes.RegionCode = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["regionCode"].(string)
	pricing.Product.Attributes.Servicecode = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["servicecode"].(string)
	pricing.Product.Attributes.LicenseModel = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["licenseModel"].(string)
	pricing.Product.Attributes.DeploymentOption = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["deploymentOption"].(string)
	pricing.Product.Attributes.Location = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["location"].(string)
	pricing.Product.Attributes.Servicename = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["servicename"].(string)
	pricing.Product.Attributes.Operation = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["operation"].(string)

	return pricing
}

// Example:
//
//	"product": {
//	  "productFamily": "Cache Instance",
//	  "attributes": {
//	    "memory": "13.07 GiB",
//	    "vcpu": "2",
//	    "instanceType": "cache.r5.large",
//	    "usagetype": "EUW3-NodeUsage:cache.r5.large",
//	    "locationType": "AWS Region",
//	    "instanceFamily": "Memory optimized",
//	    "cacheEngine": "Memcached",
//	    "regionCode": "eu-west-3",
//	    "servicecode": "AmazonElastiCache",
//	    "currentGeneration": "Yes",
//	    "networkPerformance": "Up to 10 Gigabit",
//	    "location": "EU (Paris)",
//	    "servicename": "Amazon ElastiCache",
//	    "operation": "CreateCacheCluster:0001"
//	  },
//	  "sku": "223SCNAF37X3F5SU"
//	}
func elasticachePricing(pricing *Pricing, p map[string]interface{}) *Pricing {
	pricing.Product.Attributes.Memory = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["memory"].(string)
	pricing.Product.Attributes.Vcpu = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["vcpu"].(string)
	pricing.Product.Attributes.InstanceType = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["instanceType"].(string)
	pricing.Product.Attributes.UsageType = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["usagetype"].(string)
	pricing.Product.Attributes.LocationType = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["locationType"].(string)
	pricing.Product.Attributes.InstanceFamily = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["instanceFamily"].(string)
	pricing.Product.Attributes.Engine = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["cacheEngine"].(string)
	pricing.Product.Attributes.RegionCode = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["regionCode"].(string)
	pricing.Product.Attributes.Servicecode = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["servicecode"].(string)
	pricing.Product.Attributes.CurrentGeneration = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["currentGeneration"].(string)
	pricing.Product.Attributes.NetworkPerformance = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["networkPerformance"].(string)
	pricing.Product.Attributes.Location = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["location"].(string)
	pricing.Product.Attributes.Servicename = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["servicename"].(string)
	pricing.Product.Attributes.Operation = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["operation"].(string)

	return pricing
}
