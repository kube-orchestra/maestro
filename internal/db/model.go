package db

import (
	"encoding/json"
	"fmt"
	"os"

	"gorm.io/datatypes"
)

type Consumer struct {
	Id     string         `json:"id"`
	Labels datatypes.JSON `gorm:"column:payload;type:jsonb"`
}

// had to add a T to prevent clash
type TResource struct {
	Id           string         `json:"id"`
	ConsumerId   string         `json:"consumerId"`
	GenerationId int64          `json:"generationId"`
	Object       datatypes.JSON `gorm:"column:object;type:jsonb"`
	Status       datatypes.JSON `gorm:"column:status;type:jsonb"`
}

// make sure that table name is resources and
// not t_resources
func (TResource) TableName() string {
	return "resources"
}

func updateConsumer(consumer Consumer) {
	fmt.Printf("Data inserted into DB %v\n", consumer)
	if result := dbClient.Save(&consumer); result.Error != nil {
		println("Unable to update DB with the consumer -------: : ", result.Error)
	}
}

func getConsumer(consumerId string) Consumer {
	var consumer Consumer

	if err := dbClient.Where("id = ?", consumerId).First(&consumer).Error; err != nil {
		fmt.Printf("Database select to get the consumer from the database fails -------: %v", err)
	}
	//fmt.Printf("Data being returned from DB %v\n", consumer)
	println("Data being returned from DB ", consumer.pString())
	return consumer
}

func updateResource(resource TResource) {
	fmt.Printf("Data inserted into DB %v\n", resource)
	if result := dbClient.Save(&resource); result.Error != nil {
		println("Unable to update DB with the resource -------: : ", result.Error)
	}
}

func getResource(resourceId string) TResource {
	var resource TResource

	if err := dbClient.Where("id = ?", resourceId).First(&resource).Error; err != nil {
		fmt.Printf("Database select to get the resource from the database fails -------: %v", err)
	}
	fmt.Printf("Data being returned from DB %v\n", resource)
	//println("Data being returned from DB ", consumer.pString())
	return resource
}

func updateResourceStatus(resourceId string, statusData []byte) {

	var resource TResource
	resource.Id = resourceId
	err := json.Unmarshal(statusData, &resource.Status)
	if err != nil {
		fmt.Printf("Unable to json Un Marshal gorm resource status -------: %v", err)
	}

	fmt.Printf("Status data inserted into DB %v\n", resource.Status)

	if result := dbClient.Save(&resource); result.Error != nil {
		//if result := dbClient.Save(&resource{Id: resourceId, Status: resource.Status}); result.Error != nil {
		println("Unable to update DB with the resource -------: : ", result.Error)
	}
}

func setupModel() {
	println("Setting up GORM")
	err := dbClient.AutoMigrate(&Consumer{}, &TResource{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to DB to automigrate the schema -------:: %v\n", err)
	}
}

func (consumer Consumer) pString() string {
	//return println("User: ",user.ID, " Name: ", user.Name, " Email: ",user.Email)
	return fmt.Sprintf("Id: %s Labels: %s", consumer.Id, consumer.Labels)
}
