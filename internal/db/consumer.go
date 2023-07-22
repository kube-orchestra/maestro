package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
)

const ConsumerTable = "Consumers"

func PutConsumer(c *v1.Consumer) error {
	jsonBytes, err := attributevalue.MarshalMap(c)
	if err != nil {
		return err
	}

	_, err = dbClient.PutItem(
		context.TODO(),
		&dynamodb.PutItemInput{
			TableName: aws.String(ConsumerTable),
			Item:      jsonBytes,
		})

	return err
}

func GetConsumer(consumerID string) (*v1.Consumer, error) {
	getItemInput := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{Value: consumerID},
		},
		TableName: aws.String(ConsumerTable),
	}

	c := v1.Consumer{}

	result, err := dbClient.GetItem(context.TODO(), getItemInput)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, &ErrorNotFound{}
	}

	err = attributevalue.UnmarshalMap(result.Item, &c)
	return &c, err
}
