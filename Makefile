# Makefile
CLANG ?= clang
GO := go
IMAGE_NAME ?= kubenetinsight
TAG ?= latest #$(shell date +%Y%m%d%H%M%S)
HELM_CHART_DIR := manifests/helm/kubenetinsight
EBPF_DIR := ebpf
GO_DIR := cmd/kubenetinsight
BIN_DIR := bin

.PHONY: all build test lint security-check docker-build helm-package clean

all: security-check test build docker-build helm-package

# Build targets
build: $(BIN_DIR)/kubenetinsight $(EBPF_DIR)/monitor.o

$(BIN_DIR)/kubenetinsight: $(GO_DIR)/main.go
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -ldflags="-w -s" -o $@ $<

$(EBPF_DIR)/monitor.o: $(EBPF_DIR)/monitor.c
	$(CLANG) -O2 -g -Wall -target bpf -D__TARGET_ARCH_x86 -I/usr/include -c $< -o $@

# Testing
test:
	$(GO) test -v -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...
	shellcheck scripts/*.sh

security-check:
	gosec ./...
	trivy config --severity HIGH,CRITICAL .

# Docker targets
docker-build: build
	docker buildx inspect kubenetinsight-builder >/dev/null 2>&1 || docker buildx create --use --name=kubenetinsight-builder
	docker buildx build \
		--platform linux/amd64 \
		--load \
		-t $(IMAGE_NAME):$(TAG) \
		-t $(IMAGE_NAME):latest \
		--build-arg VERSION=$(TAG) \
		.

docker-push:
	docker push $(IMAGE_NAME):$(TAG)
	docker push $(IMAGE_NAME):latest

# Helm targets
helm-package:
	helm dependency update $(HELM_CHART_DIR)
	helm package $(HELM_CHART_DIR) --app-version $(TAG)

# Deploy
deploy: #docker-build helm-package
# Build image directly in Minikube's Docker daemon
	eval $$(minikube -p minikube docker-env) && \
	docker build -t $(IMAGE_NAME):$(TAG) .
	helm upgrade --install kubenetinsight $(HELM_CHART_DIR) \
		--namespace kube-system \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(TAG)

clean:
	rm -rf $(BIN_DIR) $(EBPF_DIR)/*.o
	docker image prune -f