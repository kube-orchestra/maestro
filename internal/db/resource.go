package db

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	//v1 "github.com/kube-orchestra/maestro/proto/api/v1"
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

	var resource TResource

	body, err := json.Marshal(r)
	if err != nil {
		println("Unable to marshal json body from proto consumer -------: ", err)
		return err
	}

	_err := json.Unmarshal(body, &resource)
	if _err != nil {
		println("Unable to unmarshal json to gorm consumer -------: ", _err)
		return _err
	}
	fmt.Printf("To be inserted into the DB %v\n", resource)

	updateResource(resource)
	return nil

}

func GetResource(resourceID string) (*Resource, error) {

	tResource := getResource(resourceID)
	//fmt.Printf("Data being returned from DB %v\n", tResource)

	newResource := &Resource{}
	resourceBA, err := json.Marshal(tResource)
	if err != nil {
		fmt.Printf("Unable to json Marshal gorm resource -------: %v", err)
		return nil, err
	}

	_err := json.Unmarshal(resourceBA, newResource)
	if _err != nil {
		fmt.Printf("Unable to get json unmarshal gorm resource to proto resource -------: %v", _err)
		return nil, _err
	}
	fmt.Printf("Returning protobuf consumer from DB %v\n", newResource)

	return newResource, nil
}

func SetStatusResource(resourceID string, statusData []byte) error {
	// var status map[string]interface{}
	// if err := json.Unmarshal(statusData, &status); err != nil {
	// 	return err
	// }

	// statusAV, err := attributevalue.MarshalMap(status)
	// if err != nil {
	// 	return err
	// }

	// input := &dynamodb.UpdateItemInput{
	// 	TableName: aws.String(ResourceTable),
	// 	Key: map[string]types.AttributeValue{
	// 		"Id": &types.AttributeValueMemberS{Value: resourceID},
	// 	},
	// 	UpdateExpression: aws.String("SET #statusField = :statusValue"),
	// 	ExpressionAttributeNames: map[string]string{
	// 		"#statusField": "Status",
	// 	},
	// 	ExpressionAttributeValues: map[string]types.AttributeValue{
	// 		":statusValue": &types.AttributeValueMemberM{
	// 			Value: statusAV,
	// 		},
	// 	},
	// }

	// _, err = dbClient.UpdateItem(context.TODO(), input)
	// return err

	updateResourceStatus(resourceID, statusData)
	return nil
}
