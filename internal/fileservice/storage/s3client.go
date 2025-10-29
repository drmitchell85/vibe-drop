package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(bucket, region, endpoint string) (*S3Client, error) {
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

	// Create S3 client with custom endpoint for LocalStack
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			// Force path-style addressing (required for LocalStack)
			o.UsePathStyle = true
		}
	})

	client := &S3Client{
		client: s3Client,
		bucket: bucket,
	}

	log.Printf("S3 Client created for bucket: %s, endpoint: %s", bucket, endpoint)
	return client, nil
}

// Test connection by listing buckets
func (s *S3Client) TestConnection(ctx context.Context) error {
	_, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("failed to connect to S3: %w", err)
	}
	log.Println("S3 connection test successful")
	return nil
}

// GenerateUploadURL creates a presigned URL for uploading a file
func (s *S3Client) GenerateUploadURL(ctx context.Context, filename string) (string, string, error) {
	// Generate unique file ID
	fileID := uuid.New().String()
	key := fmt.Sprintf("%s-%s", fileID, filename)

	presignClient := s3.NewPresignClient(s.client)
	
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	
	if err != nil {
		return "", "", fmt.Errorf("failed to generate upload URL: %w", err)
	}
	
	return request.URL, fileID, nil
}

// GenerateDownloadURL creates a presigned URL for downloading a file
func (s *S3Client) GenerateDownloadURL(ctx context.Context, s3Key string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}
	
	return request.URL, nil
}

// DeleteObject deletes a file from S3
func (s *S3Client) DeleteObject(ctx context.Context, s3Key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}
	
	log.Printf("Deleted S3 object: %s", s3Key)
	return nil
}

// MultipartUploadInfo contains details for a multipart upload
type MultipartUploadInfo struct {
	UploadID string
	Key      string
}

// InitiateMultipartUpload starts a multipart upload process
func (s *S3Client) InitiateMultipartUpload(ctx context.Context, filename string) (*MultipartUploadInfo, error) {
	fileID := uuid.New().String()
	key := fmt.Sprintf("%s-%s", fileID, filename)

	result, err := s.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	info := &MultipartUploadInfo{
		UploadID: *result.UploadId,
		Key:      key,
	}

	log.Printf("Initiated multipart upload: %s (uploadID: %s)", key, info.UploadID)
	return info, nil
}

// GenerateMultipartUploadURL creates presigned URLs for each chunk
func (s *S3Client) GenerateMultipartUploadURL(ctx context.Context, uploadInfo *MultipartUploadInfo, partNumber int) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(uploadInfo.Key),
		PartNumber: aws.Int32(int32(partNumber)),
		UploadId:   aws.String(uploadInfo.UploadID),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate multipart upload URL for part %d: %w", partNumber, err)
	}

	return request.URL, nil
}