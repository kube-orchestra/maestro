#!/bin/bash

export AWS_ENDPOINT="http://localhost:8000"
export AWS_ACCESS_KEY_ID="dummy"
export AWS_SECRET_ACCESS_KEY="dummy"
export AWS_SESSION_TOKEN="dummy"
export MQTT_CLIENT_ID="maestro-api"
export MQTT_BROKER_URL="tcp://localhost:1883"
export MQTT_BROKER_USERNAME="admin"
export MQTT_BROKER_PASSWORD="password"
export MQTT_TOPIC_PREFIX="v1"

make run