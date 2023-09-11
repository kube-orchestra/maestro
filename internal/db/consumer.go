package db

import (
	"encoding/json"
	"fmt"

	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
)

const ConsumerTable = "Consumers"

func PutConsumer(c *v1.Consumer) error {

	var consumer Consumer

	body, err := json.Marshal(c)
	if err != nil {
		println("Unable to marshal json body from proto consumer -------: ", err)
		return err
	}

	_err := json.Unmarshal(body, &consumer)
	if _err != nil {
		println("Unable to unmarshal json to gorm consumer -------: ", _err)
		return _err
	}
	fmt.Printf("To be inserted into the DB %v\n", consumer)

	updateConsumer(consumer)
	return nil
}

func GetConsumer(consumerID string) (*v1.Consumer, error) {

	consumer := getConsumer(consumerID)
	newConsumer := &v1.Consumer{}
	consumerBA, err := json.Marshal(consumer)
	if err != nil {
		fmt.Printf("Unable to json Marshal gorm consumer -------: %v", err)
		return nil, err
	}

	_err := json.Unmarshal(consumerBA, newConsumer)
	if _err != nil {
		fmt.Printf("Unable to get json unmarshal gorm consumer to proto consumer -------: %v", _err)
		return nil, _err
	}
	fmt.Printf("Returning protobuf consumer from DB %v\n", newConsumer)
	return newConsumer, nil
}
