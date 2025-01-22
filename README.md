# KubeNetInsight

A real-time network monitoring solution for Kubernetes clusters using eBPF technology.

## Implemented Features

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

## Work in Progress

### Visualization
- Grafana dashboard integration
- Real-time network topology visualization
- Traffic flow analysis displays
- Custom metric dashboards

### Performance Optimization
- eBPF map optimization
- Multi-core performance tuning
- Memory usage optimization
- Metric collection efficiency

## Upcoming Features

### Advanced Analytics
- Network policy compliance monitoring
- Anomaly detection
- Traffic pattern analysis
- Historical trend analysis

### Security Features
- Network security monitoring
- Suspicious traffic detection
- Policy violation alerts
- Traffic anomaly detection

## Requirements
- Linux kernel 5.15 or later
- Kubernetes cluster (tested with Minikube)
- Go 1.23
- clang and LLVM for eBPF compilation

## Current Status
The project now features a robust metrics implementation with comprehensive Prometheus integration. Key improvements include histogram-based latency tracking, packet size distribution metrics, and detailed protocol-specific monitoring. The system provides rich insights into cluster networking through consolidated statistics and enhanced Kubernetes resource correlation.

## Connect with Me
- [GitHub](https://github.com/paras-bhavnani)
- [LinkedIn](https://www.linkedin.com/in/paras-bhavnani)
