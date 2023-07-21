.PHONY: dynamodb-start dynamodb-stop

dynamodb-start:
	docker run --rm -d -p 8000:8000 --name dynamodb  amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb
	PAGER=cat AWS_ACCESS_KEY_ID=x AWS_SECRET_ACCESS_KEY=x aws dynamodb create-table --cli-input-json file://hack/resources.table.json --region us-east-1 --endpoint-url http://localhost:8000
	PAGER=cat AWS_ACCESS_KEY_ID=x AWS_SECRET_ACCESS_KEY=x aws dynamodb create-table --cli-input-json file://hack/consumers.table.json --region us-east-1 --endpoint-url http://localhost:8000

dynamodb-stop:
	docker stop dynamodb
