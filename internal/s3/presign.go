package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"time"
)

// PreSignOptions holds parameters for generating a presigned URL.
type PreSignOptions struct {
	Presigner  *s3.PresignClient
	Bucket     string
	FilePath   string
	Expiration time.Duration
	Timeout    time.Duration
}

// PreSignOption defines a functional option for PreSignOptions.
type PreSignOption func(*PreSignOptions)

// WithPresigner sets the presign client.
func WithPresigner(presigner *s3.PresignClient) PreSignOption {
	return func(o *PreSignOptions) {
		o.Presigner = presigner
	}
}

// WithPresignBucket sets the S3 bucket.
func WithPresignBucket(bucket string) PreSignOption {
	return func(o *PreSignOptions) {
		o.Bucket = bucket
	}
}

// WithPresignFilePath sets the file path for which to generate the URL.
func WithPresignFilePath(filePath string) PreSignOption {
	return func(o *PreSignOptions) {
		o.FilePath = filePath
	}
}

// WithPresignExpiration sets the expiration duration for the URL.
func WithPresignExpiration(expiration time.Duration) PreSignOption {
	return func(o *PreSignOptions) {
		o.Expiration = expiration
	}
}

// WithPresignTimeout sets a custom timeout for generating the URL.
func WithPresignTimeout(timeout time.Duration) PreSignOption {
	return func(o *PreSignOptions) {
		o.Timeout = timeout
	}
}

// PreSignURL generates a presigned URL to get an object from S3 using the provided options.
func PreSignURL(opts ...PreSignOption) (string, error) {
	options := &PreSignOptions{
		Expiration: 15 * time.Minute, // Default expiration
		Timeout:    5 * time.Second,  // Default timeout
	}

	// Apply functional options.
	for _, opt := range opts {
		opt(options)
	}

	if options.Presigner == nil {
		return "", fmt.Errorf("presign client is required")
	}
	if options.Bucket == "" {
		return "", fmt.Errorf("s3 bucket is required")
	}
	if options.FilePath == "" {
		return "", fmt.Errorf("file path is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	request, err := options.Presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(options.Bucket),
		Key:    aws.String(options.FilePath),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = options.Expiration
	})
	if err != nil {
		return "", fmt.Errorf("unable to create presigned URL: %w", err)
	}

	return request.URL, nil
}
