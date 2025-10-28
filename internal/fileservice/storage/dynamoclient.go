package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoClient struct {
	client *dynamodb.Client
}

func NewDynamoClient(region, endpoint string) (*DynamoClient, error) {
	// For LocalStack, we need to provide fake credentials
	// In production, these would come from AWS IAM roles or environment variables
	creds := credentials.NewStaticCredentialsProvider(
		"test",      // Access Key ID (fake for LocalStack)
		"test",      // Secret Access Key (fake for LocalStack)
		"",          // Session Token (not needed)
	)

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(creds),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client with custom endpoint for LocalStack
	dynamoClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	client := &DynamoClient{
		client: dynamoClient,
	}

	log.Printf("DynamoDB Client created for region: %s, endpoint: %s", region, endpoint)
	return client, nil
}

// Test connection by listing tables
func (d *DynamoClient) TestConnection(ctx context.Context) error {
	_, err := d.client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return fmt.Errorf("failed to connect to DynamoDB: %w", err)
	}
	log.Println("DynamoDB connection test successful")
	return nil
}