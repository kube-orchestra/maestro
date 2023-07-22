package main

import (
	"context"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	client, err := mkLonglivedClient("1")
	if err != nil {
		log.Fatalln("Failed to create client:", err)
	}

	client.start()

}

// longlivedClient holds the long lived gRPC client fields
type longlivedClient struct {
	client v1.CloudEventServiceClient // client is the long lived gRPC client
	conn   *grpc.ClientConn           // conn is the client gRPC connection
	id     string                     // id is the client ID used for subscribing
}

// mkLonglivedClient creates a new client instance
func mkLonglivedClient(id string) (*longlivedClient, error) {
	conn, err := mkConnection()
	if err != nil {
		return nil, err
	}
	return &longlivedClient{
		client: v1.NewCloudEventServiceClient(conn),
		conn:   conn,
		id:     id,
	}, nil
}

// close is not used but is here as an example of how to close the gRPC client connection
func (c *longlivedClient) close() {
	if err := c.conn.Close(); err != nil {
		log.Fatal(err)
	}
}

// subscribe subscribes to messages from the gRPC server
func (c *longlivedClient) subscribe() (v1.CloudEventService_SubscribeClient, error) {
	log.Printf("Subscribing client ID: %s", c.id)
	return c.client.Subscribe(context.Background(), &v1.Request{Id: c.id})
}

// unsubscribe unsubscribes to messages from the gRPC server
func (c *longlivedClient) unsubscribe() error {
	log.Printf("Unsubscribing client ID %s", c.id)
	_, err := c.client.Unsubscribe(context.Background(), &v1.Request{Id: c.id})
	return err
}

func (c *longlivedClient) start() {
	var err error
	// stream is the client side of the RPC stream
	var stream v1.CloudEventService_SubscribeClient
	for {
		if stream == nil {
			if stream, err = c.subscribe(); err != nil {
				log.Printf("Failed to subscribe: %v", err)
				c.sleep()
				// Retry on failure
				continue
			}
		}
		response, err := stream.Recv()
		if err != nil {
			log.Printf("Failed to receive message: %v", err)
			// Clearing the stream will force the client to resubscribe on next iteration
			stream = nil
			c.sleep()
			// Retry on failure
			continue
		}
		log.Printf("Client ID %s got response: %s", c.id, response)
	}
}

// sleep is used to give the server time to unsubscribe the client and reset the stream
func (c *longlivedClient) sleep() {
	time.Sleep(time.Second * 5)
}

func mkConnection() (*grpc.ClientConn, error) {
	return grpc.Dial("localhost:8080", []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}
