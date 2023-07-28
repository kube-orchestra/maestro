package db

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ktypes "k8s.io/apimachinery/pkg/types"
)

const ResourceTable = "Resources"

type Resource struct {
	Id                   string
	ConsumerId           string
	ResourceGenerationID int64
	Object               unstructured.Unstructured
	Status               StatusMessage
}

func (r *Resource) GetUID() ktypes.UID {
	return ktypes.UID(r.Id)
}

func (r *Resource) GetResourceVersion() string {
	return strconv.FormatInt(r.ResourceGenerationID, 10)
}

func (r *Resource) GetDeletionTimestamp() *metav1.Time {
	return r.Object.GetDeletionTimestamp()
}

func (r *Resource) SetDeletionTimestamp(timestamp *metav1.Time) {
	r.Object.SetDeletionTimestamp(timestamp)
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

func ListResource() ([]*Resource, error) {
	var resources []*Resource
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(ResourceTable),
	}
	result, err := dbClient.Scan(context.TODO(), scanInput)
	if err != nil {
		return resources, err
	}

	for _, i := range result.Items {
		r := &Resource{}
		err := attributevalue.UnmarshalMap(i, r)
		if err != nil {
			return resources, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func SetStatusResource(resourceID string, status map[string]interface{}) error {
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
