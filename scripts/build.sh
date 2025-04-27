#!/bin/bash
# scripts/deploy.sh
#!/bin/bash
set -euo pipefail

# Get parameters from environment or defaults
IMAGE_NAME="${IMAGE_NAME:-ghcr.io/yourorg/kubenetinsight}"
TAG="${TAG:-$(date +%Y%m%d%H%M%S)}"
NAMESPACE="${NAMESPACE:-kube-system}"

# Build and push
make docker-build docker-push

# Apply Kubernetes manifests
helm upgrade --install kubenetinsight manifests/helm/kubenetinsight \
    --namespace "$NAMESPACE" \
    --create-namespace \
    --set image.repository="$IMAGE_NAME" \
    --set image.tag="$TAG" \
    --atomic \
    --timeout 5m

# Verify deployment
kubectl rollout status daemonset/kubenetinsight -n "$NAMESPACE" --timeout=3m