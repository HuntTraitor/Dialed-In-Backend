package s3

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"net/textproto"
)

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
func UploadToS3(client s3iface.S3API, file bytes.Buffer, fileType textproto.MIMEHeader, fileLocation string, bucket string) (string, error) {

	// generate a unique filename to store in s3
	fileName, err := generateUniqueFileName()
	if err != nil {
		return "", err
	}

	fmt.Println(fileType.Get("Content-Type"))

	// store file in s3
	output, err := client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileLocation + fileName),
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
