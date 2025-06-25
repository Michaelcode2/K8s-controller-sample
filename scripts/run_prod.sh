#!/bin/bash

# Run the Kubernetes controller in production mode
export ENV=prod

echo "Starting Kubernetes Controller in PRODUCTION mode..."
echo "Environment: $ENV"
echo "Log Level: INFO and above"
echo "Output: Structured JSON format"
echo ""

./controller controller "$@" 