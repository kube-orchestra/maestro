package main

import (
	"context"
	"log"

	v1 "github.com/kube-orchestra/maestro/proto/api/v1"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := v1.NewCloudEventServiceClient(conn)

	cloudEvent := &v1.CloudEventMessage{
		SpecVersion: "1.0",
		Type:        "com.example.someevent",
		Source:      "/mycontext",
		Id:          "A234-1234-1234",
		Time:        "2023-07-19T12:34:56Z",
		Data: map[string]string{
			"key": "value",
		},
	}

	_, err = client.SendCloudEvent(context.Background(), cloudEvent)
	if err != nil {
		log.Fatalf("could not send CloudEvent: %v", err)
	}
}
