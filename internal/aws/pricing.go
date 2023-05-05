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

type Price struct {
	Product struct {
		ProductFamily string
		Attributes    struct {
			Memory                      string
			Vcpu                        string
			InstanceType                string
			UsageType                   string
			LocationType                string
			InstanceFamily              string
			OSEngine                    string
			RegionCode                  string
			Servicecode                 string
			CurrentGeneration           string
			NetworkPerformance          string
			Location                    string
			Servicename                 string
			Operation                   string
			EngineCode                  string
			InstanceTypeFamily          string
			Storage                     string
			NormalizationSizeFactor     string
			DatabaseEdition             string
			PhysicalProcessor           string
			LicenseModel                string
			DeploymentOption            string
			ProcessorArchitecture       string
			EnhancedNetworkingSupported string
			IntelTurboAvailable         string
			DedicatedEbsThroughput      string
			Classicnetworkingsupport    string
			Capacitystatus              string
			IntelAvx2Available          string
			ClockSpeed                  string
			Ecu                         string
			GpuMemory                   string
			Vpcnetworkingsupport        string
			Tenancy                     string
			IntelAvxAvailable           string
			ProcessorFeatures           string
			PreInstalledSw              string
			Marketoption                string
			Availabilityzone            string
		}
	}
	ServiceCode         string
	OnDemandPricePerUSD string
}

func GetProducts(cfg aws.Config, serviceCode string) ([]*Price, error) {
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

	var p []*Price

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("Failed to get products: %w", err)
		}

		p, err = parsePricing(serviceCode, p, output.PriceList)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse products: %w", err)
		}

	}

	return p, nil
}

func parsePricing(serviceCode string, prices []*Price, priceList []string) ([]*Price, error) {
	for _, plist := range priceList {
		p, err := parseProduct(plist)
		if err != nil {
			return nil, err
		}

		attr := p["product"].(map[string]interface{})["attributes"].(map[string]interface{})

		// If the product does not have vcpu or memory, skip it.
		if attr["vcpu"] == nil || attr["memory"] == nil {
			continue
		}

		price := &Price{}
		price.ServiceCode = serviceCode
		price.Product.ProductFamily = p["product"].(map[string]interface{})["productFamily"].(string)

		// TODO: Looking for a better way to get the price.
		sku := p["product"].(map[string]interface{})["sku"].(string)
		skuOfferTermCode := fmt.Sprintf("%s.%s", sku, "JRTCKXETXF")
		skuOfferTermCodeRateCode := fmt.Sprintf("%s.%s.%s", sku, "JRTCKXETXF", "6YS6EN2CT7")

		// OnDemand Terms has nil data.
		if p["terms"].(map[string]interface{})["OnDemand"] == nil {
			continue
		}

		price.OnDemandPricePerUSD = p["terms"].(map[string]interface{})["OnDemand"].(map[string]interface{})[skuOfferTermCode].(map[string]interface{})["priceDimensions"].(map[string]interface{})[skuOfferTermCodeRateCode].(map[string]interface{})["pricePerUnit"].(map[string]interface{})["USD"].(string)

		switch serviceCode {
		case "AmazonEC2":
			price = price.addEc2Attributes(attr)
		case "AmazonRDS":
			price = price.addRdsAttributes(attr)
		case "AmazonElastiCache":
			price = price.addElasticacheAttributes(attr)
		default:
			panic("Unknown service code")
		}

		prices = append(prices, price)
	}

	return prices, nil
}

func parseProduct(price string) (map[string]interface{}, error) {
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(price), &product); err != nil {
		return nil, err
	}
	return product, nil
}

// product example:
//
//	"product": {
//	  "productFamily": "Compute Instance",
//	  "attributes": {
//	    "enhancedNetworkingSupported": "Yes",
//	    "intelTurboAvailable": "Yes",
//	    "memory": "4 GiB",
//	    "dedicatedEbsThroughput": "Up to 10000 Mbps",
//	    "vcpu": "2",
//	    "classicnetworkingsupport": "false",
//	    "capacitystatus": "Used",
//	    "locationType": "AWS Region",
//	    "storage": "EBS only",
//	    "instanceFamily": "Compute optimized",
//	    "operatingSystem": "Windows",
//	    "intelAvx2Available": "Yes",
//	    "regionCode": "ap-northeast-1",
//	    "physicalProcessor": "Intel Xeon 8375C (Ice Lake)",
//	    "clockSpeed": "3.5 GHz",
//	    "ecu": "NA",
//	    "networkPerformance": "Up to 12500 Megabit",
//	    "servicename": "Amazon Elastic Compute Cloud",
//	    "gpuMemory": "NA",
//	    "vpcnetworkingsupport": "true",
//	    "instanceType": "c6i.large",
//	    "tenancy": "Dedicated",
//	    "usagetype": "APN1-DedicatedUsage:c6i.large",
//	    "normalizationSizeFactor": "4",
//	    "intelAvxAvailable": "Yes",
//	    "processorFeatures": "Intel AVX; Intel AVX2; Intel AVX512; Intel Turbo",
//	    "servicecode": "AmazonEC2",
//	    "licenseModel": "No License required",
//	    "currentGeneration": "Yes",
//	    "preInstalledSw": "SQL Web",
//	    "location": "Asia Pacific (Tokyo)",
//	    "processorArchitecture": "64-bit",
//	    "marketoption": "OnDemand",
//	    "operation": "RunInstances:0202",
//	    "availabilityzone": "NA"
//	  },
//	  "sku": "2223B6PCG6QAUYY6"
//	}
func (p *Price) addEc2Attributes(attr map[string]interface{}) *Price {
	// Forcefully accommodating differences between database engines.
	if attr["enhancedNetworkingSupported"] == nil {
		attr["enhancedNetworkingSupported"] = "unknown"
	} else {
		p.Product.Attributes.EnhancedNetworkingSupported = attr["enhancedNetworkingSupported"].(string)
	}

	if attr["intelTurboAvailable"] == nil {
		attr["intelTurboAvailable"] = "unknown"
	} else {
		p.Product.Attributes.IntelTurboAvailable = attr["intelTurboAvailable"].(string)
	}

	if attr["dedicatedEbsThroughput"] == nil {
		attr["dedicatedEbsThroughput"] = "unknown"
	} else {
		p.Product.Attributes.DedicatedEbsThroughput = attr["dedicatedEbsThroughput"].(string)
	}

	if attr["intelAvx2Available"] == nil {
		attr["intelAvx2Available"] = "unknown"
	} else {
		p.Product.Attributes.IntelAvx2Available = attr["intelAvx2Available"].(string)
	}
	if attr["clockSpeed"] == nil {
		attr["clockSpeed"] = "unknown"
	} else {
		p.Product.Attributes.ClockSpeed = attr["clockSpeed"].(string)
	}

	if attr["gpuMemory"] == nil {
		attr["gpuMemory"] = "unknown"
	} else {
		p.Product.Attributes.GpuMemory = attr["gpuMemory"].(string)
	}

	if attr["intelAvxAvailable"] == nil {
		attr["intelAvxAvailable"] = "unknown"
	} else {
		p.Product.Attributes.IntelAvxAvailable = attr["intelAvxAvailable"].(string)
	}

	if attr["processorFeatures"] == nil {
		attr["processorFeatures"] = "unknown"
	} else {
		p.Product.Attributes.ProcessorFeatures = attr["processorFeatures"].(string)
	}

	p.Product.Attributes.Memory = attr["memory"].(string)
	p.Product.Attributes.Vcpu = attr["vcpu"].(string)
	p.Product.Attributes.Classicnetworkingsupport = attr["classicnetworkingsupport"].(string)
	p.Product.Attributes.Capacitystatus = attr["capacitystatus"].(string)
	p.Product.Attributes.LocationType = attr["locationType"].(string)
	p.Product.Attributes.Storage = attr["storage"].(string)
	p.Product.Attributes.InstanceFamily = attr["instanceFamily"].(string)
	p.Product.Attributes.OSEngine = attr["operatingSystem"].(string)
	p.Product.Attributes.RegionCode = attr["regionCode"].(string)
	p.Product.Attributes.PhysicalProcessor = attr["physicalProcessor"].(string)
	p.Product.Attributes.Ecu = attr["ecu"].(string)
	p.Product.Attributes.NetworkPerformance = attr["networkPerformance"].(string)
	p.Product.Attributes.Servicename = attr["servicename"].(string)
	p.Product.Attributes.Vpcnetworkingsupport = attr["vpcnetworkingsupport"].(string)
	p.Product.Attributes.InstanceType = attr["instanceType"].(string)
	p.Product.Attributes.Tenancy = attr["tenancy"].(string)
	p.Product.Attributes.UsageType = attr["usagetype"].(string)
	p.Product.Attributes.NormalizationSizeFactor = attr["normalizationSizeFactor"].(string)
	p.Product.Attributes.Servicecode = attr["servicecode"].(string)
	p.Product.Attributes.LicenseModel = attr["licenseModel"].(string)
	p.Product.Attributes.CurrentGeneration = attr["currentGeneration"].(string)
	p.Product.Attributes.PreInstalledSw = attr["preInstalledSw"].(string)
	p.Product.Attributes.Location = attr["location"].(string)
	p.Product.Attributes.ProcessorArchitecture = attr["processorArchitecture"].(string)
	p.Product.Attributes.Marketoption = attr["marketoption"].(string)
	p.Product.Attributes.Operation = attr["operation"].(string)
	p.Product.Attributes.Availabilityzone = attr["availabilityzone"].(string)

	return p
}

// product example:
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
func (p *Price) addRdsAttributes(attr map[string]interface{}) *Price {
	// Forcefully accommodating differences between database engines.
	if attr["engineCode"] == nil {
		p.Product.Attributes.EngineCode = "unknown"
	} else {
		p.Product.Attributes.EngineCode = attr["engineCode"].(string)
	}

	if attr["databaseEdition"] == nil {
		p.Product.Attributes.DatabaseEdition = "unknown"
	} else {
		p.Product.Attributes.DatabaseEdition = attr["databaseEdition"].(string)
	}

	if attr["physicalProcessor"] == nil {
		p.Product.Attributes.PhysicalProcessor = "unknown"
	} else {
		p.Product.Attributes.PhysicalProcessor = attr["physicalProcessor"].(string)
	}

	if attr["currentGeneration"] == nil {
		p.Product.Attributes.CurrentGeneration = "unknown"
	} else {
		p.Product.Attributes.CurrentGeneration = attr["currentGeneration"].(string)
	}

	if attr["networkPerformance"] == nil {
		p.Product.Attributes.NetworkPerformance = "unknown"
	} else {
		p.Product.Attributes.NetworkPerformance = attr["networkPerformance"].(string)
	}

	if attr["processorArchitecture"] == nil {
		p.Product.Attributes.ProcessorArchitecture = "unknown"
	} else {
		p.Product.Attributes.ProcessorArchitecture = attr["processorArchitecture"].(string)
	}

	p.Product.Attributes.InstanceTypeFamily = attr["instanceTypeFamily"].(string)
	p.Product.Attributes.Memory = attr["memory"].(string)
	p.Product.Attributes.Vcpu = attr["vcpu"].(string)
	p.Product.Attributes.InstanceType = attr["instanceType"].(string)
	p.Product.Attributes.UsageType = attr["usagetype"].(string)
	p.Product.Attributes.LocationType = attr["locationType"].(string)
	p.Product.Attributes.Storage = attr["storage"].(string)
	p.Product.Attributes.NormalizationSizeFactor = attr["normalizationSizeFactor"].(string)
	p.Product.Attributes.InstanceFamily = attr["instanceFamily"].(string)
	p.Product.Attributes.OSEngine = attr["databaseEngine"].(string)
	p.Product.Attributes.RegionCode = attr["regionCode"].(string)
	p.Product.Attributes.Servicecode = attr["servicecode"].(string)
	p.Product.Attributes.LicenseModel = attr["licenseModel"].(string)
	p.Product.Attributes.DeploymentOption = attr["deploymentOption"].(string)
	p.Product.Attributes.Location = attr["location"].(string)
	p.Product.Attributes.Servicename = attr["servicename"].(string)
	p.Product.Attributes.Operation = attr["operation"].(string)

	return p
}

// product example:
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
func (p *Price) addElasticacheAttributes(attr map[string]interface{}) *Price {
	p.Product.Attributes.Memory = attr["memory"].(string)
	p.Product.Attributes.Vcpu = attr["vcpu"].(string)
	p.Product.Attributes.InstanceType = attr["instanceType"].(string)
	p.Product.Attributes.UsageType = attr["usagetype"].(string)
	p.Product.Attributes.LocationType = attr["locationType"].(string)
	p.Product.Attributes.InstanceFamily = attr["instanceFamily"].(string)
	p.Product.Attributes.OSEngine = attr["cacheEngine"].(string)
	p.Product.Attributes.RegionCode = attr["regionCode"].(string)
	p.Product.Attributes.Servicecode = attr["servicecode"].(string)
	p.Product.Attributes.CurrentGeneration = attr["currentGeneration"].(string)
	p.Product.Attributes.NetworkPerformance = attr["networkPerformance"].(string)
	p.Product.Attributes.Location = attr["location"].(string)
	p.Product.Attributes.Servicename = attr["servicename"].(string)
	p.Product.Attributes.Operation = attr["operation"].(string)

	return p
}
