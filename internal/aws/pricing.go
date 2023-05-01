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
			Vcpu         string
			Memory       string
			InstanceType string
			Engine       string
			RegionCode   string
			Operation    string
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
		pricing.Product.Attributes.Vcpu = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["vcpu"].(string)
		pricing.Product.Attributes.Memory = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["memory"].(string)
		pricing.Product.Attributes.InstanceType = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["instanceType"].(string)
		pricing.Product.Attributes.Engine = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["cacheEngine"].(string)
		pricing.Product.Attributes.RegionCode = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["regionCode"].(string)
		pricing.Product.Attributes.Operation = p["product"].(map[string]interface{})["attributes"].(map[string]interface{})["operation"].(string)

		// TODO: Looking for a better way to get the price.
		sku := p["product"].(map[string]interface{})["sku"].(string)
		skuOfferTermCode := fmt.Sprintf("%s.%s", sku, "JRTCKXETXF")
		skuOfferTermCodeRateCode := fmt.Sprintf("%s.%s.%s", sku, "JRTCKXETXF", "6YS6EN2CT7")

		pricing.PricePerUSD = p["terms"].(map[string]interface{})["OnDemand"].(map[string]interface{})[skuOfferTermCode].(map[string]interface{})["priceDimensions"].(map[string]interface{})[skuOfferTermCodeRateCode].(map[string]interface{})["pricePerUnit"].(map[string]interface{})["USD"].(string)

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
