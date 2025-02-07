# Performance Optimization Runbook

## Resource Management

### Monitoring Commands

```bash
kubectl top pods
kubectl top nodes
```

### Optimization Steps

1. Resource Allocation
   - Set appropriate CPU/memory limits
   - Configure HPA
   - Implement resource quotas

2. Network Performance
   - Monitor network metrics
   - Optimize CNI settings
   - Configure network policies
