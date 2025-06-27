#!/bin/bash

# Run the Kubernetes controller in development mode
export ENV=dev

echo "Starting Kubernetes Controller in DEVELOPMENT mode..."
echo "Environment: $ENV"
echo "Log Level: DEBUG"
echo "Output: Pretty console with emojis"
echo ""

./controller controller "$@"