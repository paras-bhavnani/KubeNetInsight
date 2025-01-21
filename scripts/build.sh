#!/bin/bash

# Generate tag using timestamp
TAG=$(date +%Y%m%d%H%M%S)
IMAGE_NAME="kubenetinsight"

echo "Building new version with tag: $TAG"

# Build and load the image
docker build -t ${IMAGE_NAME}:${TAG} -f Dockerfile .
minikube image load ${IMAGE_NAME}:${TAG}

# Update the DaemonSet
kubectl set image daemonset/${IMAGE_NAME} ${IMAGE_NAME}=${IMAGE_NAME}:${TAG} -n kube-system

# Force a rollout
kubectl rollout restart daemonset/${IMAGE_NAME} -n kube-system

# Clean up old Docker images
echo "Cleaning up old Docker images..."
docker images ${IMAGE_NAME} --format "{{.ID}} {{.Tag}}" | sort -k2 -r | tail -n +4 | awk '{print $1}' | xargs -r docker rmi

# Clean up old Minikube images
echo "Cleaning up old Minikube images..."
minikube ssh "docker images ${IMAGE_NAME} --format '{{.ID}} {{.Tag}}' | sort -k2 -r | tail -n +4 | awk '{print \$1}' | xargs -r docker rmi"

# Remove images with the 'latest' tag
echo "Removing images with 'latest' tag..."
docker images ${IMAGE_NAME}:latest --format "{{.ID}}" | xargs -r docker rmi
minikube ssh "docker images ${IMAGE_NAME}:latest --format '{{.ID}}' | xargs -r docker rmi"

# Remove dangling images
echo "Removing dangling images..."
docker image prune -f
minikube ssh "docker image prune -f"

echo "Updated to version $TAG and cleaned up old images"