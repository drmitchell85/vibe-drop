package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoClient struct {
	client *dynamodb.Client
}

// FileMetadata represents the structure for file metadata in DynamoDB
type FileMetadata struct {
	FileID      string `json:"fileID" dynamodbav:"fileID"`
	Filename    string `json:"filename" dynamodbav:"filename"`
	TotalSize   int64  `json:"totalSize" dynamodbav:"totalSize"`
	ContentType string `json:"contentType" dynamodbav:"contentType"`
	Status      string `json:"status" dynamodbav:"status"`
	UploadType  string `json:"uploadType" dynamodbav:"uploadType"`
	UploadedAt  string `json:"uploadedAt" dynamodbav:"uploadedAt"`
	UserID      string `json:"userID" dynamodbav:"userID"`
	S3Key       string `json:"s3Key" dynamodbav:"s3Key"`
	// Future chunking fields (will be empty for single uploads)
	S3UploadID   *string `json:"s3UploadId,omitempty" dynamodbav:"s3UploadId,omitempty"`
	ChunkSize    *int64  `json:"chunkSize,omitempty" dynamodbav:"chunkSize,omitempty"`
	TotalChunks  *int    `json:"totalChunks,omitempty" dynamodbav:"totalChunks,omitempty"`
	CompletedAt  *string `json:"completedAt,omitempty" dynamodbav:"completedAt,omitempty"`
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

// SaveFileMetadata saves file metadata to DynamoDB
func (d *DynamoClient) SaveFileMetadata(ctx context.Context, metadata *FileMetadata) error {
	// Convert struct to DynamoDB item
	item, err := attributevalue.MarshalMap(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Save to DynamoDB
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("vibe-drop-files"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save file metadata: %w", err)
	}

	log.Printf("Saved file metadata for fileID: %s", metadata.FileID)
	return nil
}

// GetFileMetadata retrieves file metadata by fileID
func (d *DynamoClient) GetFileMetadata(ctx context.Context, fileID string) (*FileMetadata, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("vibe-drop-files"),
		Key: map[string]types.AttributeValue{
			"fileID": &types.AttributeValueMemberS{Value: fileID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("file not found: %s", fileID)
	}

	var metadata FileMetadata
	err = attributevalue.UnmarshalMap(result.Item, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// ListUserFiles retrieves all files for a specific user
func (d *DynamoClient) ListUserFiles(ctx context.Context, userID string) ([]FileMetadata, error) {
	// For now, we'll scan the entire table and filter by userID
	// In production, this would use a GSI on userID
	result, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String("vibe-drop-files"),
		FilterExpression: aws.String("userID = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list user files: %w", err)
	}

	var files []FileMetadata
	for _, item := range result.Items {
		var metadata FileMetadata
		err = attributevalue.UnmarshalMap(item, &metadata)
		if err != nil {
			log.Printf("Failed to unmarshal item: %v", err)
			continue
		}
		files = append(files, metadata)
	}

	return files, nil
}

// DeleteFileMetadata removes file metadata from DynamoDB
func (d *DynamoClient) DeleteFileMetadata(ctx context.Context, fileID string) error {
	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("vibe-drop-files"),
		Key: map[string]types.AttributeValue{
			"fileID": &types.AttributeValueMemberS{Value: fileID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	log.Printf("Deleted file metadata for fileID: %s", fileID)
	return nil
}

// FileChunk represents a single chunk in the chunks table
type FileChunk struct {
	FileID      string `json:"fileID" dynamodbav:"fileID"`
	ChunkNumber int    `json:"chunkNumber" dynamodbav:"chunkNumber"`
	Size        int64  `json:"size" dynamodbav:"size"`
	ETag        string `json:"etag" dynamodbav:"etag"`
	Status      string `json:"status" dynamodbav:"status"` // "pending", "uploaded", "failed"
	UploadedAt  string `json:"uploadedAt,omitempty" dynamodbav:"uploadedAt,omitempty"`
	S3PartNumber int   `json:"s3PartNumber" dynamodbav:"s3PartNumber"`
}

// SaveFileChunk saves chunk metadata to DynamoDB
func (d *DynamoClient) SaveFileChunk(ctx context.Context, chunk *FileChunk) error {
	item, err := attributevalue.MarshalMap(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("vibe-drop-chunks"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save chunk metadata: %w", err)
	}

	log.Printf("Saved chunk metadata for fileID: %s, chunk: %d", chunk.FileID, chunk.ChunkNumber)
	return nil
}

// GetFileChunks retrieves all chunks for a file
func (d *DynamoClient) GetFileChunks(ctx context.Context, fileID string) ([]FileChunk, error) {
	result, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String("vibe-drop-chunks"),
		KeyConditionExpression: aws.String("fileID = :fileID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fileID": &types.AttributeValueMemberS{Value: fileID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get chunks: %w", err)
	}

	var chunks []FileChunk
	for _, item := range result.Items {
		var chunk FileChunk
		err = attributevalue.UnmarshalMap(item, &chunk)
		if err != nil {
			log.Printf("Failed to unmarshal chunk item: %v", err)
			continue
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// UpdateChunkStatus updates a chunk's upload status and ETag
func (d *DynamoClient) UpdateChunkStatus(ctx context.Context, fileID string, chunkNumber int, status string, etag string) error {
	updateExpression := "SET #status = :status"
	expressionAttributeNames := map[string]string{
		"#status": "status",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{Value: status},
	}

	// Add ETag and uploadedAt if status is "uploaded"
	if status == "uploaded" && etag != "" {
		updateExpression += ", etag = :etag, uploadedAt = :uploadedAt"
		expressionAttributeValues[":etag"] = &types.AttributeValueMemberS{Value: etag}
		expressionAttributeValues[":uploadedAt"] = &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}
	}

	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("vibe-drop-chunks"),
		Key: map[string]types.AttributeValue{
			"fileID":      &types.AttributeValueMemberS{Value: fileID},
			"chunkNumber": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", chunkNumber)},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if err != nil {
		return fmt.Errorf("failed to update chunk status: %w", err)
	}

	log.Printf("Updated chunk %d status to %s for fileID: %s", chunkNumber, status, fileID)
	return nil
}

// CheckUploadComplete checks if all chunks are uploaded and returns completion status
func (d *DynamoClient) CheckUploadComplete(ctx context.Context, fileID string) (bool, []FileChunk, error) {
	chunks, err := d.GetFileChunks(ctx, fileID)
	if err != nil {
		return false, nil, err
	}

	// Check if all chunks are uploaded
	for _, chunk := range chunks {
		if chunk.Status != "uploaded" {
			return false, chunks, nil // Not complete yet
		}
	}

	return len(chunks) > 0, chunks, nil // Complete if we have chunks and all are uploaded
}