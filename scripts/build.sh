#!/bin/bash

set -e

echo "Building ORCA..."

# Change to backend directory
cd backend

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Build main server
echo "Building ORCA server..."
go build -o orca ./cmd/orca

# Build CLI
echo "Building ORCA CLI..."
go build -o orca-cli ./cmd/orca-cli

echo "Build complete!"
echo "Binaries:"
echo "  - backend/orca (server)"
echo "  - backend/orca-cli (CLI tool)"