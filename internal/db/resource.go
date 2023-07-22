package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const ResourceTable = "Resources"

type Resource struct {
	Id                   string
	ConsumerId           string
	ResourceGenerationID int64
	Object               unstructured.Unstructured
}

func PutResource(r *Resource) error {
	jsonBytes, err := attributevalue.MarshalMap(r)
	if err != nil {
		return err
	}

	_, err = dbClient.PutItem(
		context.TODO(),
		&dynamodb.PutItemInput{
			TableName: aws.String(ResourceTable),
			Item:      jsonBytes,
		})

	return err
}

func GetResource(resourceID string) (*Resource, error) {
	getItemInput := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{Value: resourceID},
		},
		TableName: aws.String(ResourceTable),
	}

	r := Resource{}

	result, err := dbClient.GetItem(context.TODO(), getItemInput)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, &ErrorNotFound{}
	}

	err = attributevalue.UnmarshalMap(result.Item, &r)
	return &r, err
}
