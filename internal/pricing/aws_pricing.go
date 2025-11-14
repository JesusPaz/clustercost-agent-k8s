package pricing

import (
	"fmt"
)

// NodePriceResolver exposes node level pricing.
type NodePriceResolver interface {
	// NodeHourlyPrice returns the on-demand hourly price in USD for the provided instance type.
	NodeHourlyPrice(instanceType string) (float64, bool)
}

// AWSPricing resolves prices from a static configuration map.
type AWSPricing struct {
	region string
	prices map[string]float64
}

// NewAWSPricing builds a resolver for a region using the provided configuration map.
func NewAWSPricing(region string, regionPrices map[string]map[string]float64) (*AWSPricing, error) {
	if region == "" {
		return nil, fmt.Errorf("aws pricing requires a region")
	}
	priceMap := regionPrices[region]
	if priceMap == nil {
		priceMap = map[string]float64{}
	}
	return &AWSPricing{
		region: region,
		prices: priceMap,
	}, nil
}

// NodeHourlyPrice implements the NodePriceResolver interface.
func (a *AWSPricing) NodeHourlyPrice(instanceType string) (float64, bool) {
	price, ok := a.prices[instanceType]
	return price, ok
}
