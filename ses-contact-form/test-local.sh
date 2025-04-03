#!/bin/bash
set -e

# Build the Go binary for Lambda
echo "Building Go binary..."
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Check if build succeeded
if [ $? -ne 0 ]; then
    echo "Build failed"
    exit 1
fi

# Run the Lambda function locally using SAM
echo "Invoking Lambda function locally using SAM..."
sam local invoke ContactFormFunction -e sam-events/test-event.json

# Clean up
echo "Cleaning up..."
rm bootstrap

echo "Local test completed!"
