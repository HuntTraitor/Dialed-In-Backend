package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/textproto"
	"time"
)

type UploadOptions struct {
	Client      *s3.Client
	File        bytes.Buffer
	FileType    textproto.MIMEHeader
	FilePath    string
	Bucket      string
	OldFileName string
	Timeout     time.Duration
}

// Option is a function that modifies UploadOptions
type Option func(options *UploadOptions)

// WithClient sets the S3 client
func WithClient(client *s3.Client) Option {
	return func(o *UploadOptions) {
		o.Client = client
	}
}

// WithFile sets the file buffer
func WithFile(file bytes.Buffer) Option {
	return func(o *UploadOptions) {
		o.File = file
	}
}

// WithFileType sets the file type header
func WithFileType(fileType textproto.MIMEHeader) Option {
	return func(o *UploadOptions) {
		o.FileType = fileType
	}
}

// WithFilePath sets the file path
func WithFilePath(filePath string) Option {
	return func(o *UploadOptions) {
		o.FilePath = filePath
	}
}

// WithBucket sets the S3 bucket
func WithBucket(bucket string) Option {
	return func(o *UploadOptions) {
		o.Bucket = bucket
	}
}

// WithOldFileName sets an existing file name (if updating an existing file)
func WithOldFileName(oldFileName string) Option {
	return func(o *UploadOptions) {
		o.OldFileName = oldFileName
	}
}

// WithTimeout sets a custom timeout
func WithTimeout(timeout time.Duration) Option {
	return func(o *UploadOptions) {
		o.Timeout = timeout
	}
}

// UploadToS3 uploads a file to S3 and returns the unique filename
func UploadToS3(opts ...Option) (string, error) {
	options := &UploadOptions{
		Timeout: 5 * time.Second, // Default timeout
	}

	// Apply functional options
	for _, opt := range opts {
		opt(options)
	}

	if options.Client == nil {
		return "", fmt.Errorf("s3 client is required")
	}
	if options.Bucket == "" {
		return "", fmt.Errorf("s3 bucket is required")
	}

	// Generate a unique filename if not provided
	fileName := options.OldFileName
	if fileName == "" {
		var err error
		fileName, err = generateUniqueFileName()
		if err != nil {
			return "", err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	// Upload to S3
	output, err := options.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(options.Bucket),
		Key:         aws.String(options.FilePath + fileName),
		Body:        bytes.NewReader(options.File.Bytes()),
		ContentType: aws.String(options.FileType.Get("Content-Type")),
	})
	if err != nil {
		return "", err
	}

	// Ensure upload was successful
	if output.ETag == nil {
		return "", fmt.Errorf("upload may have failed, missing ETag")
	}

	return fileName, nil
}
