.PHONY: run stop

run:
	#docker run --rm -d -p 8000:8000 --name dynamodb  amazon/dynamodb-local
	docker run --rm -d -p 8000:8000 --name dynamodb  amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb

delete:
	docker stop dynamodb

tables:
	aws dynamodb delete-table --table-name Resources --region us-east-1 --endpoint-url http://localhost:8000 || :
	aws dynamodb create-table --cli-input-json file://resources.table.json --region us-east-1 --endpoint-url http://localhost:8000

	aws dynamodb delete-table --table-name Consumers --region us-east-1 --endpoint-url http://localhost:8000 || :
	aws dynamodb create-table --cli-input-json file://consumers.table.json --region us-east-1 --endpoint-url http://localhost:8000

