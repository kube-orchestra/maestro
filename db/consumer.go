package db

import (
	"context"
	"fmt"
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

func ListConsumers() (*v1.ConsumerList, error) {

	input := &dynamodb.ScanInput{
		TableName: aws.String(ConsumerTable),
	}

	var consumers []*v1.Consumer

	paginator := dynamodb.NewScanPaginator(dbClient, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return &v1.ConsumerList{}, fmt.Errorf("failed to retrieve items from DynamoDB: %v", err)
		}

		for _, item := range output.Items {
			consumer := v1.Consumer{
				Id:     item["Id"].(*types.AttributeValueMemberS).Value,
				Name:   item["Name"].(*types.AttributeValueMemberS).Value,
				Labels: []*v1.ConsumerLabel{},
			}

			// Process the labels if present
			if labelsAttr, ok := item["Labels"]; ok {
				labels := make([]*v1.ConsumerLabel, len(labelsAttr.(*types.AttributeValueMemberL).Value))
				for i, labelAttr := range labelsAttr.(*types.AttributeValueMemberL).Value {
					label := v1.ConsumerLabel{
						Key:   labelAttr.(*types.AttributeValueMemberM).Value["Key"].(*types.AttributeValueMemberS).Value,
						Value: labelAttr.(*types.AttributeValueMemberM).Value["Value"].(*types.AttributeValueMemberS).Value,
					}
					labels[i] = &label
				}
				consumer.Labels = labels
			}

			consumers = append(consumers, &consumer)
		}
	}

	consumerList := &v1.ConsumerList{
		Consumers: consumers,
	}
	return consumerList, nil

}
