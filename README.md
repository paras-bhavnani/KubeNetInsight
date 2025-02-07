# KubeNetInsight

A real-time network monitoring solution for Kubernetes clusters using eBPF technology.

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

### Enhanced Metrics Capabilities
- Real-time metric querying and analysis
- Detailed latency tracking with percentile distributions:
  - Average latency measurements
  - 10th, 50th, 90th, and 99th percentile analysis
  - Source/destination IP correlation
- Network traffic analysis:
  - Byte-level traffic monitoring
  - Time-series data collection
  - Per-connection statistics
- Connection state tracking:
  - TCP connection states (ESTABLISHED, etc.)
  - Protocol-specific monitoring
  - Port-level details

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

### RAG (Retrieval Augmented Generation)
- Vector database implementation using FAISS
- Document embedding generation and storage
- Optimized semantic search functionality
- REST API for document retrieval
- Comprehensive runbooks covering:
  - Pod lifecycle management
  - Network troubleshooting
  - Performance optimization
  - Monitoring and alerts
  - KubeNetInsight-specific issues

## Requirements
- Linux kernel 5.15 or later
- Kubernetes cluster (tested with Minikube)
- Go 1.23
- clang and LLVM for eBPF compilation
- Prometheus and Grafana for metrics visualization

## Project Structure
```
.
├── cmd/
│   └── kubenetinsight/
│       └── main.go        # Main application entry point and network monitoring logic
├── pkg/
│   ├── api/
│   │   └── search.py      # Search API implementation
│   ├── embeddings/
│   │   ├── model.py       # Embedding model implementation
│   │   ├── index.py       # FAISS index management
│   │   └── optimizer.py   # Query optimization
│   ├── ebpf/
│   │   ├── monitor.c      # eBPF program for packet capture and analysis
│   │   └── monitor.o      # Compiled eBPF object file
│   ├── kubernetes/
│   │   └── client.go      # Kubernetes API client integration
│   └── metrics/
│       ├── __init__.py    # Python package initialization
│       ├── exporter.go    # Prometheus metrics exporter implementation
│       └── prometheus_client.py  # Python client for querying Prometheus metrics
├── manifests/
│   ├── documentation/
│   │   └── runbooks/      # Operational runbooks and troubleshooting guides
│   └── monitoring/
│       ├── grafana/
│       │   ├── dashboards/  # Grafana dashboard configurations
│       │   ├── deployment.yaml  # Grafana deployment configuration
│       │   └── secret.yaml  # Grafana secrets configuration
│       ├── kubenetinsight/
│       │   └── metrics-service.yaml  # Metrics service configuration
│       └── prometheus/
│           └── prometheus-deployment.yaml  # Prometheus deployment and config
├── scripts/
│   └── build.sh          # Build automation script
├── .gitignore           # Git ignore patterns
├── Dockerfile           # Container image build configuration
├── Makefile            # Build and deployment automation
├── README.md           # Project documentation
├── daemonset.yaml      # DaemonSet deployment configuration
├── go.mod              # Go module dependencies
├── go.sum              # Go module checksums
├── kubenetinsight-role.yaml        # RBAC role configuration
├── kubenetinsight-rolebinding.yaml # RBAC role binding configuration
└── kubenetinsight-sa.yaml          # Service account configuration
```

## Current Status
The project now features a robust metrics implementation with comprehensive Prometheus and Grafana integration. Key improvements include histogram-based latency tracking, packet size distribution metrics, and detailed protocol-specific monitoring. The system provides rich insights into cluster networking through consolidated statistics and enhanced Kubernetes resource correlation. The latest implementation includes detailed connection tracking with proper endianness handling and visualization capabilities through custom Grafana dashboards.

Recent enhancements include:
- Advanced latency analysis with percentile-based tracking
- Comprehensive network traffic monitoring with time-series data
- Detailed connection state tracking and protocol analysis
- Enhanced metric querying capabilities for real-time analysis

## Connect with Me
- [GitHub](https://github.com/paras-bhavnani)
- [LinkedIn](https://www.linkedin.com/in/paras-bhavnani)
