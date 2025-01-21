# KubeNetInsight

A real-time network monitoring solution for Kubernetes clusters using eBPF technology.

## Implemented Features

### eBPF Network Monitoring
- Kernel-level packet capture using XDP (eXpress Data Path)
- Real-time packet monitoring across multiple CPU cores
- Comprehensive packet capture and analysis with the following capabilities:
  - Source and destination IP tracking
  - Packet count monitoring
  - Packet size tracking
  - Connection tracking with source/destination ports
  - Protocol-specific metrics (TCP/UDP)
  - Latency measurements
  - Packet drop monitoring
  - Multi-core processing support

### Core Infrastructure
- Integration with Kubernetes cluster (Minikube)
- eBPF program loading and attachment
- Metrics collection system
- Real-time packet capture and analysis

### Kubernetes Integration
- DaemonSet creation for cluster-wide deployment
- Enhanced pod and service discovery
- Namespace-aware monitoring
- IP address correlation with Kubernetes resources
- Traffic mapping to Kubernetes services

### Metrics and Monitoring
- Basic Prometheus metrics exporter implementation
- Custom metrics for:
  - Network traffic (packet counts and bytes)
  - Connection latency
  - Packet drops
  - Protocol-specific counts

### Data Processing and Visualization
- Consolidated network traffic summary
- Detailed connection information
- Summary statistics including total packets, bytes, unique sources/destinations, and protocol breakdown

## Work in Progress

### Metrics and Monitoring
- Advanced Prometheus metrics
- Custom metrics refinement

### Visualization
- Grafana dashboard integration
- Real-time network topology visualization
- Traffic flow analysis displays

### Performance Optimization
- eBPF map optimization
- Multi-core performance tuning
- Memory usage optimization

## Upcoming Features

### Advanced Analytics
- Network policy compliance monitoring
- Anomaly detection
- Traffic pattern analysis

### Security Features
- Network security monitoring
- Suspicious traffic detection
- Policy violation alerts

## Requirements
- Linux kernel 5.15 or later
- Kubernetes cluster (tested with Minikube)
- Go 1.23
- clang and LLVM for eBPF compilation

## Current Status
The project has successfully implemented comprehensive network monitoring capabilities, including packet tracking, latency measurements, and Kubernetes resource correlation. It now features cluster-wide deployment via DaemonSet, enhanced pod and service discovery, and namespace-aware monitoring. The system is actively being developed to include advanced features for more in-depth Kubernetes network insights.

## Connect with Me
- [GitHub](https://github.com/paras-bhavnani)
- [LinkedIn](https://www.linkedin.com/in/paras-bhavnani)
