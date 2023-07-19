.PHONY: run stop

run:
	#docker run --rm -d -p 8000:8000 --name dynamodb  amazon/dynamodb-local
	docker run --rm -d -p 8000:8000 --name dynamodb  amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb

delete:
	docker stop dynamodb
