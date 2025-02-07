# Pod Lifecycle Runbook

## Pod Creation Failures

### Symptoms

- ImagePullBackOff status
- Pending state
- CrashLoopBackOff

### Diagnostic Commands

```bash
kubectl get pods -n <namespace>
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

### Resolution Steps

1. Image Pull Errors
   - Verify image name in deployment YAML
   - Check container registry access
   - Validate image tag

2. Resource Constraints
   - Review resource requests/limits
   - Check node capacity
   - Verify resource quotas

## Pod Crashes

### Diagnostic Steps

1. Check container logs
2. Review resource usage
3. Analyze restart patterns
