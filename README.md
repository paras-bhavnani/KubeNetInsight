# KubeNetInsight

A real-time network monitoring solution for Kubernetes clusters using eBPF technology.

## Quickstart

### Prerequisites
- Kubernetes cluster (e.g. Minikube)
- Helm 3
- Docker or Minikube Docker daemon
- Go 1.23, clang/LLVM (for local builds)

### Deploy with Helm

```
Build and deploy locally in Minikube
make deploy

or manually:
eval $(minikube docker-env)
docker build -t kubenetinsight:latest .
helm upgrade --install kubenetinsight manifests/helm/kubenetinsight
--namespace kube-system --create-namespace
--set image.pullPolicy=Never
```

### Verify

```
kubectl get daemonset,service -n kube-system -l app.kubernetes.io/name=kubenetinsight
kubectl port-forward svc/kubenetinsight-metrics 8080:8080 -n kube-system
curl localhost:8080/metrics
```

## Features

### eBPF Network Monitoring
- Kernel-level packet capture using XDP (eXpress Data Path)
- Real-time packet monitoring across multiple CPU cores
- Comprehensive packet capture and analysis with the following capabilities:
  - Source and destination IP tracking
  - Packet count monitoring
  - Packet size tracking with distribution metrics
  - Connection tracking with source/destination ports
  - Protocol-specific metrics (TCP/UDP)
  - Latency measurements with histogram metrics
  - Packet drop monitoring
  - Multi-core processing support

### Core Infrastructure
- Integration with Kubernetes cluster (Minikube)
- eBPF program loading and attachment
- Consolidated metrics collection system
- Real-time packet capture and analysis
- Enhanced data processing pipeline

### Kubernetes Integration
- DaemonSet creation for cluster-wide deployment
- Enhanced pod and service discovery
- Namespace-aware monitoring
- IP address correlation with Kubernetes resources
- Traffic mapping to Kubernetes services
- Resource count metrics per namespace

### Metrics and Monitoring
- Comprehensive Prometheus metrics exporter implementation
- Custom metrics for:
  - Network traffic (packet counts and bytes)
  - Connection latency histograms
  - Packet size distributions
  - Protocol-specific traffic
  - Connection states
  - Packet drops
  - Pod and service counts per namespace

### Data Processing and Visualization
- Consolidated network traffic summary
- Detailed connection information with protocol states
- Rich summary statistics including:
  - Total packets and bytes
  - Unique sources/destinations
  - Protocol breakdown
  - Connection states
  - Latency distributions
- Grafana dashboards for:
  - Network topology visualization
  - Traffic pattern analysis
  - Performance metrics monitoring
  - Alert configuration

## Requirements
- Linux kernel 5.15 or later
- Kubernetes cluster (tested with Minikube)
- Go 1.23
- clang and LLVM for eBPF compilation
- Helm 3 
- Prometheus and Grafana for metrics visualization

## Project Structure
```
.
├── cmd/kubenetinsight/         # Main application entry point and network monitoring logic
├── ebpf/                       # XDP/eBPF C program source
├── manifests/                  # Helm chart for deployment
│   ├── helm/
│   │    └── kubenetinsight/
│   │        ├── Chart.yaml
│   │        ├── values.yaml
│   │        └── templates/     # DaemonSet, RBAC, Service, ConfigMap, Prometheus config
├── pkg/
│   ├── ebpf/                   # eBPF program and collector
│   ├── kubernetes/             # Kubernetes client integration
│   └── metrics/                # Prometheus metrics exporter
├── scripts/                    # Build and deployment scripts for testing/dev
│   └── build.sh
├── Makefile                    # Build, Docker, and Helm automation
├── Dockerfile                  # Secure multi-stage container build
└──  .dockerignore              # Exclude unneeded files from Docker context
```

## Usage
- `make deploy` to build image in Minikube and deploy via Helm  
- `make clean` to remove local build artifacts 

## Current Status
The project now features a robust metrics implementation with comprehensive Prometheus and Grafana integration. Key improvements include histogram-based latency tracking, packet size distribution metrics, and detailed protocol-specific monitoring. The system provides rich insights into cluster networking through consolidated statistics and enhanced Kubernetes resource correlation. The latest implementation includes detailed connection tracking with proper endianness handling and visualization capabilities through custom Grafana dashboards.

## Connect with Me
- [GitHub](https://github.com/paras-bhavnani)
- [LinkedIn](https://www.linkedin.com/in/paras-bhavnani)
