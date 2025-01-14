# KubeNetInsight

A real-time network monitoring solution for Kubernetes clusters using eBPF technology.

## Implemented Features

### eBPF Network Monitoring
- Kernel-level packet capture using XDP (eXpress Data Path)
- Real-time packet monitoring across multiple CPU cores
- Successful packet capture and analysis with the following capabilities:
  - Source and destination IP tracking
  - Packet count monitoring
  - Multi-core processing support (demonstrated across cores 004-011)

### Core Infrastructure
- Integration with Kubernetes cluster (Minikube)
- eBPF program loading and attachment
- Basic metrics collection system
- Real-time packet capture verification

## Work in Progress

### Kubernetes Integration
- DaemonSet creation for cluster-wide deployment
- Pod and service discovery implementation
- Network policy integration

### Metrics and Monitoring
- Prometheus metrics exporter implementation
- Custom metrics definition for:
  - Network latency
  - Packet drops
  - Connection tracking

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
- Go 1.21+
- clang and LLVM for eBPF compilation

## Current Status
The project successfully implements basic network monitoring capabilities and is actively being developed to include advanced features for comprehensive Kubernetes network insights.

## Connect with Me
- [GitHub](https://github.com/paras-bhavnani)
- [LinkedIn](https://www.linkedin.com/in/paras-bhavnani)