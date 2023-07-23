#!/bin/bash

docker build . -t quay.io/apahim/maestro-api:v5 -f config/maestro-api.Dockerfile
docker push quay.io/apahim/maestro-api:v5
