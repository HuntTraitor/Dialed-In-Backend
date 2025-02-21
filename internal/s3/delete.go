package s3

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"log"
	"time"
)

// DeleteOptions holds parameters for deleting an S3 object.
type DeleteOptions struct {
	Client   *s3.Client
	Bucket   string
	FilePath string
	Timeout  time.Duration
}

// DeleteOption defines a functional option for DeleteOptions.
type DeleteOption func(*DeleteOptions)

// WithDeleteClient sets the S3 client.
func WithDeleteClient(client *s3.Client) DeleteOption {
	return func(o *DeleteOptions) {
		o.Client = client
	}
}

// WithDeleteBucket sets the S3 bucket.
func WithDeleteBucket(bucket string) DeleteOption {
	return func(o *DeleteOptions) {
		o.Bucket = bucket
	}
}

// WithDeleteFilePath sets the file path of the object to delete.
func WithDeleteFilePath(filePath string) DeleteOption {
	return func(o *DeleteOptions) {
		o.FilePath = filePath
	}
}

// WithDeleteTimeout sets a custom timeout for the deletion.
func WithDeleteTimeout(timeout time.Duration) DeleteOption {
	return func(o *DeleteOptions) {
		o.Timeout = timeout
	}
}

// DeleteFromS3 deletes an object from S3 using the provided options.
// It returns a boolean indicating deletion success and an error if any.
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html
func DeleteFromS3(opts ...DeleteOption) (bool, error) {
	// Set default values.
	options := &DeleteOptions{
		Timeout: 5 * time.Second,
	}

	// Apply functional options.
	for _, opt := range opts {
		opt(options)
	}

	if options.Client == nil {
		return false, fmt.Errorf("s3 client is required")
	}
	if options.Bucket == "" {
		return false, fmt.Errorf("s3 bucket is required")
	}
	if options.FilePath == "" {
		return false, fmt.Errorf("file path is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(options.Bucket),
		Key:    aws.String(options.FilePath),
	}

	_, err := options.Client.DeleteObject(ctx, input)
	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError
		if errors.As(err, &noKey) {
			log.Printf("Object %s does not exist in %s.\n", options.FilePath, options.Bucket)
			err = noKey
		} else if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				log.Printf("Access denied: cannot delete object %s from %s.\n", options.FilePath, options.Bucket)
				// Consider this non-fatal if you want deletion to be idempotent.
				err = nil
			}
		}
		// If we got an error from deletion, return false.
		if err != nil {
			return false, err
		}
	}

	// Wait for the object to be deleted.
	err = s3.NewObjectNotExistsWaiter(options.Client).Wait(
		ctx, &s3.HeadObjectInput{Bucket: aws.String(options.Bucket), Key: aws.String(options.FilePath)}, time.Minute)
	if err != nil {
		log.Printf("Failed attempt to wait for object %s in bucket %s to be deleted.\n", options.FilePath, options.Bucket)
		return false, err
	}

	return true, nil
}
