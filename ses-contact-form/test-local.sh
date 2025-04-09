#!/bin/bash
set -e

# Check if event file was provided as an argument
SAM_EVENT=${1:-"sam-events/success-event.json"}

# Build the Go binary for Lambda
echo "Building Go binary..."
GOOS=linux GOARCH=amd64 go build -o bootstrap

# Check if build succeeded
if [ $? -ne 0 ]; then
    echo "Build failed"
    exit 1
fi

# Run the Lambda function locally using SAM with mock SES client
echo "Invoking Lambda function locally using SAM CLI..."
echo "Running event file: $SAM_EVENT..."
sam local invoke ContactFormFunction -e "$SAM_EVENT"

# Clean up
echo "Cleaning up..."
rm bootstrap

echo "Local test completed!"