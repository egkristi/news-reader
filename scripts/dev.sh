#!/bin/bash

# Exit on any error
set -e

# Configuration
CONTAINER_NAME="news-reader"
IMAGE_NAME="news-reader:dev"
VOLUME_NAME="news-reader-data"
PORT="8082"
VERSION="0.1.0"
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Function to rebuild and restart the container
rebuild_and_restart() {
    echo "🔨 Building application..."
    podman build -t $IMAGE_NAME \
        --build-arg VERSION=$VERSION \
        --build-arg GIT_COMMIT=$GIT_COMMIT \
        --build-arg BUILD_TIME=$BUILD_TIME \
        .

    echo "🛑 Stopping existing container..."
    podman stop $CONTAINER_NAME 2>/dev/null || true
    podman rm $CONTAINER_NAME 2>/dev/null || true

    echo "🚀 Starting new container..."
    podman run -d \
        --name $CONTAINER_NAME \
        -p $PORT:$PORT \
        -v $VOLUME_NAME:/data \
        $IMAGE_NAME

    echo "✅ Container started successfully!"
    echo "📱 Application is available at http://localhost:$PORT"
}

# Function to run tests
run_tests() {
    echo "🧪 Running tests..."
    go test -v ./...
}

# Check if fswatch is installed
if ! command -v fswatch &> /dev/null; then
    echo "❌ fswatch is not installed. Installing..."
    brew install fswatch
fi

# Initial build and start
rebuild_and_restart

# Watch for changes
echo "👀 Watching for changes..."
fswatch -o . | while read f; do
    # Ignore certain directories and files
    if [[ $f == *"/build/"* ]] || \
       [[ $f == *"/.git/"* ]] || \
       [[ $f == *"/scripts/"* ]] || \
       [[ $f == *".swp" ]] || \
       [[ $f == *"~" ]]; then
        continue
    fi

    echo "🔄 Change detected in $f"
    
    # Run tests first
    if ! run_tests; then
        echo "❌ Tests failed, not deploying"
        continue
    fi

    # If tests pass, rebuild and restart
    rebuild_and_restart
done
