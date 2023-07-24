package db

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var dbClient *dynamodb.Client

const (
	awsEndpoint        = "AWS_ENDPOINT"
	awsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

type ErrorNotFound struct{}

func (e *ErrorNotFound) Error() string {
	return fmt.Sprintf("Resource not found")
}

func init() {
	dbClient, _ = newClient()
}

// newClient Creates a DynamoDB Client
func newClient() (*dynamodb.Client, error) {

	accessKeyID := os.Getenv(awsAccessKeyID)
	if len(accessKeyID) == 0 {
		return nil, fmt.Errorf("%s must be set", awsAccessKeyID)
	}

	secretAccessKey := os.Getenv(awsSecretAccessKey)
	if len(secretAccessKey) == 0 {
		return nil, fmt.Errorf("%s must be set", awsSecretAccessKey)
	}

	//endpoint := os.Getenv(awsEndpoint)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		//config.WithEndpointResolver(aws.EndpointResolverFunc(
		//	func(service, region string) (aws.Endpoint, error) {
		//		return aws.Endpoint{URL: fmt.Sprintf(endpoint)}, nil
		//	})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			},
		}),
	)

	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}
