package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func Config(profile, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("Failed to load AWS config: %v", err)
	}

	return cfg, nil
}
