package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	cloudeventsv1 "github.com/kube-orchestra/maestro/internal/services/v1/cloudevents"
	consumerv1 "github.com/kube-orchestra/maestro/internal/services/v1/consumers"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
)

func main() {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Register reflection services on gRPC server.
	reflection.Register(s)

	var consumersAPI = consumerv1.NewConsumerService()
	v1.RegisterConsumerServiceServer(s, consumersAPI)

	var cloudEventsAPI = cloudeventsv1.NewCloudEventsService()
	v1.RegisterCloudEventServiceServer(s, cloudEventsAPI)

	// Serve gRPC server
	log.Println("Serving gRPC on 0.0.0.0:8080")
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:8080",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = v1.RegisterConsumerServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	// mount a path to expose the generated OpenAPI specification on disk
	mux.HandleFunc("/swagger-ui/consumer.swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./swagger/api/v1/consumer.swagger.json")
	})

	// mount a path to expose the generated OpenAPI specification on disk
	mux.HandleFunc("/swagger-ui/resource.swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./swagger/api/v1/resource.swagger.json")
	})

	// mount the Swagger UI that uses the OpenAPI specification path above
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./swagger-ui"))))

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	err = http.ListenAndServe("localhost:8090", mux)
	if err != nil {
		log.Fatal(err)
	}
}
