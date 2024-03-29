.PHONY: dynamodb-start dynamodb-stop mosquitto-start mosquitto-stop

dynamodb-start:
	docker run --rm -d -p 8000:8000 --name dynamodb  amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb
	PAGER=cat AWS_ACCESS_KEY_ID=x AWS_SECRET_ACCESS_KEY=x aws dynamodb create-table --cli-input-json file://hack/resources.table.json --region us-east-1 --endpoint-url http://localhost:8000
	PAGER=cat AWS_ACCESS_KEY_ID=x AWS_SECRET_ACCESS_KEY=x aws dynamodb create-table --cli-input-json file://hack/consumers.table.json --region us-east-1 --endpoint-url http://localhost:8000

dynamodb-stop:
	docker stop dynamodb

mosquitto-start:
	docker run -d --rm -it -p 1883:1883 --name mosquitto -v $(shell pwd)/hack/mosquitto-passwd.txt:/mosquitto/config/password.txt -v $(shell pwd)/hack/mosquitto.conf:/mosquitto/config/mosquitto.conf eclipse-mosquitto

mosquitto-stop:
	docker stop mosquitto

run:
	buf generate
	go run $(shell pwd)/cmd/server/main.go