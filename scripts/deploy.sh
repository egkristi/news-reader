#!/bin/bash

# Exit on any error
set -e

# Configuration
CONTAINER_NAME="news-reader"
IMAGE_NAME="ghcr.io/$GITHUB_REPOSITORY:latest"
VOLUME_NAME="news-reader-data"
PORT="8082"

# Pull the latest image
echo "Pulling latest image..."
podman pull $IMAGE_NAME

# Stop and remove existing container
echo "Stopping existing container..."
podman stop $CONTAINER_NAME || true
podman rm $CONTAINER_NAME || true

# Create volume if it doesn't exist
echo "Ensuring volume exists..."
podman volume inspect $VOLUME_NAME >/dev/null 2>&1 || podman volume create $VOLUME_NAME

# Start new container
echo "Starting new container..."
podman run -d \
  --name $CONTAINER_NAME \
  -p $PORT:$PORT \
  -v $VOLUME_NAME:/data \
  $IMAGE_NAME

# Check if container is running
echo "Checking container status..."
if podman ps | grep -q $CONTAINER_NAME; then
  echo "Container started successfully!"
  echo "Application is available at http://localhost:$PORT"
else
  echo "Error: Container failed to start"
  podman logs $CONTAINER_NAME
  exit 1
fi
