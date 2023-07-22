package db

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

const ConsumerTable = "Consumers"

type Consumer struct {
	Id     string
	Name   string
	Labels []ConsumerLabel
	Object map[string]interface{} // To hold the google.protobuf.Struct data
}

type ConsumerLabel struct {
	Key   string
	Value string
}

func ConvertProtobufToStruct(protoConsumer *v1.Consumer) Consumer {
	consumer := Consumer{
		Id:   protoConsumer.Id,
		Name: protoConsumer.Name,
	}

	for _, protoLabel := range protoConsumer.Labels {
		label := ConsumerLabel{
			Key:   protoLabel.Key,
			Value: protoLabel.Value,
		}
		consumer.Labels = append(consumer.Labels, label)
	}

	object, err := structpb.ToGoValue(protoConsumer.Object)
	if err != nil {
		return err
	}

	if objMap, ok := object.(map[string]interface{}); ok {
		consumer.Object = objMap
	}

	return consumer
}

func PutConsumer(c Consumer) error {

	labels := make([]types.AttributeValue, len(c.Labels))
	for i, label := range c.Labels {
		labels[i] = types.AttributeValue{
			M: map[string]types.AttributeValue{
				"Key":   &types.AttributeValueMemberS{Value: label.Key},
				"Value": &types.AttributeValueMemberS{Value: label.Value},
			},
		}
	}
	objectAttr, err := structpb.NewValue(c.Object)
	objectMap, err := types.MarshalAttributeValueMap(objectAttr.GetStructValue())

	item := map[string]types.AttributeValue{
		"id":     &types.AttributeValueMemberS{Value: c.Id},
		"name":   &types.AttributeValueMemberS{Value: c.Name},
		"labels": &types.AttributeValueMemberL{Value: labels},
		"object": &types.AttributeValueMemberM{Value: objectMap},
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(ConsumerTable),
	}

	_, err = dbClient.PutItem(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil

	//jsonBytes, err := attributevalue.MarshalMap(c)
	//if err != nil {
	//	return err
	//}
	//
	//_, err = dbClient.PutItem(
	//	context.TODO(),
	//	&dynamodb.PutItemInput{
	//		TableName: aws.String(ConsumerTable),
	//		Item:      jsonBytes,
	//	})

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
