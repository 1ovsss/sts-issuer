package sts

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"sts-issuer/internal/envs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type PolicyStatement struct {
	Sid       string   `json:"Sid"`
	Effect    string   `json:"Effect"`
	Action    []string `json:"Action"`
	Principal string   `json:"Principal"`
	Resource  []string `json:"Resource"`
}

type Policy struct {
	Version    string            `json:"Version"`
	Statements []PolicyStatement `json:"Statement"`
}

var stsClient *sts.Client

// createPolicy generates a policy with the specified effect, actions, and resources.
func createPolicyFromEnv(identifier string) (string, error) {
	policy := Policy{
		Version:    "2012-10-17",
		Statements: []PolicyStatement{},
	}

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "STS_POLICY_"+identifier) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				return "", fmt.Errorf("invalid environment variable format: %s", env)
			}
			var policyStatement struct {
				Effect   string   `json:"Effect"`
				Action   []string `json:"Action"`
				Resource []string `json:"Resource"`
			}
			err := json.Unmarshal([]byte(parts[1]), &policyStatement)
			if err != nil {
				return "", err
			}
			policy.Statements = append(policy.Statements, PolicyStatement{
				Sid:       "All",
				Effect:    policyStatement.Effect,
				Action:    policyStatement.Action,
				Principal: "*", // Principal is always "*"
				Resource:  policyStatement.Resource,
			})
		}
	}

	jsonPolicy, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(jsonPolicy), nil
}

// loadAWSConfig loads the AWS configuration with a custom STS endpoint.
func loadAWSConfig(stsURL string) (aws.Config, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{URL: stsURL}, nil
	})

	return config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(resolver))
}

func InitCfg() error {
	// Load configuration
	yandexCloudSTSURL := envs.GetEnvOrDefault("YC_STS_URL", "https://sts.yandexcloud.net/")
	cfg, err := loadAWSConfig(yandexCloudSTSURL)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	// Create STS client
	stsClient = sts.NewFromConfig(cfg)
	return nil
}
