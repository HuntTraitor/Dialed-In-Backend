package s3

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"log"
	"net/textproto"
	"time"
)

type S3 struct {
	Client    *s3.Client
	Presigner *s3.PresignClient
}

func generateUniqueFileName() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

// UploadToS3 takes a client, file, and bucket and uploads a file to that s3 bucket
// and returns the unique filename identifier
func UploadToS3(client *s3.Client, file bytes.Buffer, fileType textproto.MIMEHeader, filePath string, bucket string) (string, error) {
	// generate a unique filename to store in s3
	fileName, err := generateUniqueFileName()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// store file in s3
	output, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(filePath + fileName),
		Body:        bytes.NewReader(file.Bytes()),
		ContentType: aws.String(fileType.Get("Content-Type")),
	})
	if err != nil {
		return "", err
	}
	if output.ETag == nil {
		return "", fmt.Errorf("upload may have failed, missing ETag")
	}
	return fileName, nil
}

// DeleteFromS3 deletes a filePath string form an S3 bucket
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html
func DeleteFromS3(client *s3.Client, bucket string, filePath string) (bool, error) {
	deleted := false
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filePath),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.DeleteObject(ctx, input)
	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError
		if errors.As(err, &noKey) {
			log.Printf("Object %s does not exist in %s.\n", filePath, bucket)
			err = noKey
		} else if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				log.Printf("Access denied: cannot delete object %s from %s.\n", filePath, bucket)
				err = nil
			}
		}
	} else {
		err = s3.NewObjectNotExistsWaiter(client).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(filePath)}, time.Minute)
		if err != nil {
			log.Printf("Failed attempt to wait for object %s in bucket %s to be deleted.\n", filePath, bucket)
		} else {
			deleted = true
		}
	}
	return deleted, err
}

// PreSignURL generates a presigned URL to get an object from S3 with a specified expiration time.
func PreSignURL(presigner *s3.PresignClient, bucket string, filePath string, expiration time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Generate the presigned URL with the specified expiration
	request, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filePath),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration // Set the expiration time for the presigned URL
	})

	if err != nil {
		return "", fmt.Errorf("unable to create presigned URL: %w", err)
	}

	return request.URL, nil
}
