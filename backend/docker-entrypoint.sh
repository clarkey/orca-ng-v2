#!/bin/sh
set -e

# Generate go.sum if it doesn't exist
if [ ! -f go.sum ]; then
    echo "Generating go.sum..."
    go mod tidy
fi

# Run the application
exec go run ./cmd/orca