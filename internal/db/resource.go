package db

import (
	"context"
	"encoding/json"

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
	Status               StatusMessage
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

func SetStatusResource(resourceID string, statusData []byte) error {
	var status map[string]interface{}
	if err := json.Unmarshal(statusData, &status); err != nil {
		return err
	}

	statusAV, err := attributevalue.MarshalMap(status)
	if err != nil {
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(ResourceTable),
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{Value: resourceID},
		},
		UpdateExpression: aws.String("SET #statusField = :statusValue"),
		ExpressionAttributeNames: map[string]string{
			"#statusField": "Status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":statusValue": &types.AttributeValueMemberM{
				Value: statusAV,
			},
		},
	}

	_, err = dbClient.UpdateItem(context.TODO(), input)
	return err
}
