# Network Troubleshooting Runbook

## Service Discovery Issues

### Symptoms

- Service unreachable
- DNS resolution failures
- Endpoint connection timeouts

### Diagnostic Commands

```bash
kubectl get services
kubectl get endpoints
kubectl describe service <service-name>
```

### Resolution Steps

1. Service Configuration
   - Verify service selectors
   - Check port mappings
   - Validate service type

2. DNS Troubleshooting
   - Test DNS resolution
   - Check CoreDNS pods
   - Verify network policies
