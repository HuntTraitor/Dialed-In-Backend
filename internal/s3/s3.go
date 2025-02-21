package s3

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
