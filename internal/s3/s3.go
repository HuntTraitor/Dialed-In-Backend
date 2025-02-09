package s3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// UploadToS3 takes a client, file, filename, and bucket and uploads a file to that s3 bucket
func UploadToS3(client s3iface.S3API, file []byte, filename string, bucket string) error {
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(file),
		ACL:    aws.String("public-read"),
	})
	return err
}
