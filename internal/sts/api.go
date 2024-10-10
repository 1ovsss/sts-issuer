package sts

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type STSData struct {
	Title     string   `json:"title"`
	ExpiresIn string   `json:"expires_in"`
	ARN       string   `json:"arn"`
	Policies  []string `json:"policies"`
}

type Creds struct {
	AccessKeyID     string `json:"access_key"`
	SecretAccessKey string `json:"secret_key"`
	SessionToken    string `json:"session_token"`
	Expiration      string `json:"expiration"`
}

// getSTSData scans environment variables and extracts relevant data for a given identifier
func GetSTSData(identifier string) (STSData, error) {
	var data STSData
	var policies []string

	// Loop through all environment variables
	for _, env := range os.Environ() {
		// Split the env variable into key and value
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key, value := pair[0], pair[1]

		// Match the environment variables based on the identifier
		if strings.HasPrefix(key, "STS_EXPIRES_IN_"+identifier) {
			data.ExpiresIn = value
		} else if strings.HasPrefix(key, "STS_ARN_"+identifier) {
			data.ARN = value
		} else if strings.HasPrefix(key, "STS_POLICY_"+identifier) {
			policies = append(policies, value)
		}
	}

	if data.ARN == "" || data.ExpiresIn == "" {
		return STSData{}, fmt.Errorf("missing required env variables for identifier: %s", identifier)
	}

	data.Policies = policies
	return data, nil
}

// getAllSTSData scans environment variables and finds all STS data grouped by unique identifiers
func GetAllSTSData() map[string]STSData {
	identifiers := make(map[string]STSData)

	// Loop through all environment variables to find unique identifiers
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := pair[0]

		// Extract identifier from STS_ARN_, STS_EXPIRES_IN_, or STS_POLICY_ keys
		var identifier string
		if strings.HasPrefix(key, "STS_EXPIRES_IN_") {
			identifier = strings.TrimPrefix(key, "STS_EXPIRES_IN_")
		} else if strings.HasPrefix(key, "STS_ARN_") {
			identifier = strings.TrimPrefix(key, "STS_ARN_")
		} else if strings.HasPrefix(key, "STS_POLICY_") {
			parts := strings.Split(key, "_")
			if len(parts) >= 3 {
				identifier = parts[2]
			}
		}

		// Skip if no identifier found
		if identifier == "" {
			continue
		}

		// Check if identifier already exists, if not initialize it
		if _, exists := identifiers[identifier]; !exists {
			stsData, err := GetSTSData(identifier)
			if err == nil {
				identifiers[identifier] = stsData
			}
		}
	}

	return identifiers
}

func GetCreds(identifier string) (*Creds, error) {
	// Get role parameters based on the identifier
	durationSecondsStr := os.Getenv("STS_EXPIRES_IN_" + identifier)
	if durationSecondsStr == "" {
		durationSecondsStr = "900" // Default value
	}

	durationSeconds, err := strconv.ParseInt(durationSecondsStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid durationSeconds: %v", err)
	}

	roleARN := os.Getenv("STS_ARN_" + identifier)
	if roleARN == "" {
		return nil, fmt.Errorf("missing role ARN for identifier: %s", identifier)
	}

	roleSessionName := fmt.Sprintf("session-%d", time.Now().UnixNano())

	// Create policy based on the identifier
	policy, err := createPolicyFromEnv(identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy: %v", err)
	}
	log.Println(policy)

	// Assume role
	input := &sts.AssumeRoleInput{
		DurationSeconds: aws.Int32(int32(durationSeconds)),
		Policy:          aws.String(policy),
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String(roleSessionName),
	}

	result, err := stsClient.AssumeRole(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %v", err)
	}

	log.Println("Sts creds claimed seccsessfully")

	// Return STS credentials
	return &Creds{
		AccessKeyID:     *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		SessionToken:    *result.Credentials.SessionToken,
		Expiration:      formatExpiration(*result.Credentials.Expiration),
	}, nil
}

// formatExpiration formats a time.Time value to the UTC+3 timezone in "YYYY-MM-DD HH:MM:SS" format.
func formatExpiration(expiration time.Time) string {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatal("Failed to load timezone: ", err)
	}
	return expiration.In(loc).Format("2006-01-02 15:04:05")
}
