package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kube-orchestra/maestro/internal/mqtt"
	consumerv1 "github.com/kube-orchestra/maestro/internal/service/v1/consumers"
	resourcesv1 "github.com/kube-orchestra/maestro/internal/service/v1/resources"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const listenAddress = "0.0.0.0:8080"
const listenAddressGateway = "0.0.0.0:8090"

func main() {
	ctx := context.TODO()
	mqttConnection := mqtt.NewConnection(ctx)
	mqttConnection.StartSender(ctx)
	mqttConnection.StartStatusReceiver(ctx)

	// gRPC config

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Register reflection service on gRPC server.
	reflection.Register(s)

	// Attach the consumers service to the server
	var consumersAPI = consumerv1.NewConsumerService()
	v1.RegisterConsumerServiceServer(s, consumersAPI)

	// Attach the resources service to the server
	var resourcesAPI = resourcesv1.NewResourceService(mqttConnection.ResourceChannel)
	v1.RegisterResourceServiceServer(s, resourcesAPI)

	// Serve gRPC server
	log.Println("Serving gRPC on", listenAddress)
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		"localhost:8080",
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
		log.Fatalln("Failed to register consumer service handler:", err)
	}

	err = v1.RegisterResourceServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register resource service handler:", err)
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

	log.Println("Serving gRPC-Gateway on", listenAddressGateway)
	err = http.ListenAndServe(listenAddressGateway, mux)
	if err != nil {
		log.Fatal(err)
	}

}
