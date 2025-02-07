# Monitoring and Alerts Runbook

## Prometheus Configuration

### Alert Rules

```yaml
groups:
- name: kubeNetInsight
  rules:
  - alert: HighLatency
    expr: histogram_quantile(0.95, 
          rate(http_request_duration_seconds_bucket{
            job="kubeNetInsight"
          }[5m])) > 0.5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: High latency detected
```

### Grafana Dashboards

1. Pod Status Dashboard
2. Network Traffic Dashboard
3. Latency Metrics Dashboard
