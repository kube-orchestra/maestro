# Maestro

Maestro is a component of the Kube Orchestra Project, a multi-cluster resources orchestrator for Kubernetes.

Maestro is the API for cluster registration and single-cluster resources definition.

## Kube Orchestra Architecture

![Kube Orchestra Architecture](./architecture.png)

## Run

```
$ go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
$ go install google.golang.org/protobuf/cmd/protoc-gen-go
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
$ go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
$ buf generate
$ go run cmd/server/main.go
```

## Develop

### Mosquitto

Mosquitto is an open-source MQTT broker

```shell
# Start mosquitto
make mosquitto-start

# Stop mosquitto
make mosquitto-stop
```

In order to connect to this Mosquitto server use user: `admin`, password: `password`, on port 1883. [MQTT Explorer](http://mqtt-explorer.com/) is a good client for local inspection and manipulation of the MQTT messages.

### DynamoDB

```shell
# Start local DynamoDB
make dynamodb-start

# Read necessary env to be able to use `aws` CLI
source hack/dynamodb-local-env.sh

# List created tables
aws dynamodb list-tables

# Dump all resources
aws dynamodb scan --table-name Resources
```

### Consumer

```shell
# Create a new Consumer
curl -X POST  localhost:8090/v1/consumers -H "Content-Type: application/json" -d '{"name": "Test", "labels": [{"key": "k1", "value": "v1" }]}'
{
  "id": "af467f701-f6af-408b-ba2f-9436896be890",
  "name": "Test",
  "labels": [
    {
      "key": "k1",
      "value": "v1"
    }
  ]
}

# And another one
curl -X POST  localhost:8090/v1/consumers -H "Content-Type: application/json" -d '{"name": "Test2", "labels": [{"key": "k1", "value": "v1" }]}'
{
  "id": "c497f701-f6af-408b-ba2f-9436896be537",
  "name": "Test",
  "labels": [
    {
      "key": "k1",
      "value": "v2"
    }
  ]
}

# Get a specific Consumer
curl localhost:8090/v1/consumers/c497f701-f6af-408b-ba2f-9436896be537 | jq
{
  "id": "c497f701-f6af-408b-ba2f-9436896be537",
  "name": "Test",
  "labels": [
    {
      "key": "k1",
      "value": "v2"
    }
  ]
}

```

### Integrating with ConcertMaster

```shell
# Get a Consumer ID
CONSUMER_ID=$(curl -s -X POST localhost:8090/v1/consumers|jq -r .id)

# From ConcertMaster's README
...
CONCERTMASTER_TOPIC_PREFIX=v1/ CONCERTMASTER_CLIENT_ID=$CONSUMER_ID go run ./cmd/concertmaster
...

# Create a resource
RESOURCE_ID=$(curl -s -X POST localhost:8090/v1/resources/$CONSUMER_ID -H "Content-Type: application/json" --data-binary @hack/example.deployment.json|jq -r .id)

# Update a resource
curl -X PUT localhost:8090/v1/resources/$RESOURCE_ID -H "Content-Type: application/json" --data-binary @hack/example.deployment.v2.json
```
