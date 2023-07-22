package cloudevents

import (
	"context"
	"fmt"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"sync"
)

type CloudEventsService struct {
	v1.UnimplementedCloudEventServiceServer
	subscribers sync.Map
}

func NewCloudEventsService() *CloudEventsService {
	return &CloudEventsService{}
}

func (s *CloudEventsService) SendCloudEvent(_ context.Context, c *v1.CloudEventMessage) (*emptypb.Empty, error) {
	//key := 1 // Replace "your-key" with the key you want to retrieve
	value, found := s.subscribers.Load("1")

	if found {
		// The item exists in the sync.Map
		// You can access the value using 'value'
		// For example, if your value is of type 'sub', you can do:
		subValue := value.(sub)
		// Now you can work with 'subValue' which is of type 'sub'
		// ...
		subValue.stream.Send(c)

	} else {
		// The item doesn't exist in the sync.Map
		// Handle the case where the key is not present
		// ...
		fmt.Println("key not found")
	}

	return &emptypb.Empty{}, nil
}

type sub struct {
	stream   v1.CloudEventService_SubscribeServer
	finished chan<- bool
}

func (s *CloudEventsService) Subscribe(request *v1.Request, stream v1.CloudEventService_SubscribeServer) error {
	// Handle subscribe request
	log.Printf("Received subscribe request from ID: %s", request.Id)

	fin := make(chan bool)
	// Save the subscriber stream according to the given client ID
	s.subscribers.Store(request.Id, sub{stream: stream, finished: fin})

	//key := 1 // Replace "your-key" with the key you want to retrieve
	value, found := s.subscribers.Load(request.Id)

	if found {
		// The item exists in the sync.Map
		// You can access the value using 'value'
		// For example, if your value is of type 'sub', you can do:
		subValue := value.(sub)
		// Now you can work with 'subValue' which is of type 'sub'
		// ...
		subValue.stream.Send(&v1.CloudEventMessage{
			SpecVersion: "1.0",
			Type:        "com.example.someevent",
			Source:      "/mycontext",
			Id:          "A234-1234-1234",
			Time:        "2023-07-19T12:34:56Z",
			Data: map[string]string{
				"Subscribed": request.Id,
			},
		})

	} else {
		// The item doesn't exist in the sync.Map
		// Handle the case where the key is not present
		// ...
		fmt.Println("key not found")
	}

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			log.Printf("Closing stream for client ID: %s", request.Id)
			return nil
		case <-ctx.Done():
			log.Printf("Client ID %s has disconnected", request.Id)
			return nil
		}
	}
}

func (s *CloudEventsService) Unsubscribe(ctx context.Context, request *v1.Request) (*v1.Response, error) {
	v, ok := s.subscribers.Load(request.Id)
	if !ok {
		return nil, fmt.Errorf("failed to load subscriber key: %d", request.Id)
	}
	sub, ok := v.(sub)
	if !ok {
		return nil, fmt.Errorf("failed to cast subscriber value: %T", v)
	}
	select {
	case sub.finished <- true:
		log.Printf("Unsubscribed client: %d", request.Id)
	default:
		// Default case is to avoid blocking in case client has already unsubscribed
	}
	s.subscribers.Delete(request.Id)
	return &v1.Response{}, nil
}
