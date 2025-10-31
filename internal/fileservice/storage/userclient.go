package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// User represents a user account in the system
type User struct {
	UserID       string `json:"user_id" dynamodbav:"userID"`
	Username     string `json:"username" dynamodbav:"username"`
	Email        string `json:"email" dynamodbav:"email"`
	PasswordHash string `json:"-" dynamodbav:"passwordHash"` // Never expose in JSON responses
	CreatedAt    string `json:"created_at" dynamodbav:"createdAt"`
	UpdatedAt    string `json:"updated_at" dynamodbav:"updatedAt"`
}

// CreateUser saves a new user to DynamoDB
func (d *DynamoClient) CreateUser(ctx context.Context, user *User) error {
	// Set timestamps
	now := time.Now().Format(time.RFC3339)
	user.CreatedAt = now
	user.UpdatedAt = now

	// Convert struct to DynamoDB item
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Save to DynamoDB with condition to prevent duplicate userIDs
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String("vibe-drop-users"),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(userID)"),
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by their ID
func (d *DynamoClient) GetUserByID(ctx context.Context, userID string) (*User, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("vibe-drop-users"),
		Key: map[string]types.AttributeValue{
			"userID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email address
func (d *DynamoClient) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	// Query using the email GSI
	result, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("vibe-drop-users"),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("user not found with email: %s", email)
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (d *DynamoClient) UpdateUser(ctx context.Context, user *User) error {
	// Update timestamp
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	// Convert struct to DynamoDB item
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Update in DynamoDB
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("vibe-drop-users"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}