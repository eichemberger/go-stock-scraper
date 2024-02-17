package customAWS

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/eichemberger/go-stock-scraper/src/logger"
)

var s3Client *s3.Client

func getS3Client() *s3.Client {
	if s3Client != nil {
		return s3Client
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Sugar.Fatalw("Unable to load SDK config",
			"error", err,
		)
	}
	s3Client = s3.NewFromConfig(cfg)

	return s3.NewFromConfig(cfg)
}

func S3PutObject(objBytes []byte, bucketName, objectKey string) error {
	s3Client := getS3Client()

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(objBytes),
	})

	return err
}
