# Perplexity MCP Helm Chart

Kubernetes deployment for the Perplexity MCP Server using Helm.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Perplexity API key

## Installing the Chart

### From source

```bash
# Create a values file with your API key
cat > my-values.yaml <<EOF
perplexity:
  apiKey: "pplx-xxxxxxxxxxxx"

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: perplexity-mcp.example.com
      paths:
        - path: /
          pathType: Prefix
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/auth-type: basic
    nginx.ingress.kubernetes.io/auth-secret: basic-auth
  tls:
    - secretName: perplexity-mcp-tls
      hosts:
        - perplexity-mcp.example.com
EOF

# Install the chart
helm install perplexity-mcp ./deploy/chart -f my-values.yaml
```

### Using existing secret

For production deployments, create a Kubernetes secret first:

```bash
# Create secret
kubectl create secret generic perplexity-api-key \
  --from-literal=api-key=pplx-xxxxxxxxxxxx

# Install chart referencing the secret
helm install perplexity-mcp ./deploy/chart \
  --set perplexity.existingSecret=perplexity-api-key
```

## Configuration

### Key Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `perplexity.apiKey` | Perplexity API key (not recommended for production) | `""` |
| `perplexity.existingSecret` | Name of existing secret with `api-key` field | `""` |
| `server.mode` | Server mode: `stdio` or `http` | `http` |
| `server.listenAddr` | HTTP listen address | `:8080` |
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `ghcr.io/mikluko/perplexity-mcp` |
| `image.tag` | Image tag (defaults to chart appVersion) | `""` |
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |
| `autoscaling.enabled` | Enable HPA | `false` |
| `autoscaling.minReplicas` | Minimum replicas | `1` |
| `autoscaling.maxReplicas` | Maximum replicas | `5` |

### Security Configuration

#### Basic Authentication

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/auth-type: basic
    nginx.ingress.kubernetes.io/auth-secret: basic-auth
    nginx.ingress.kubernetes.io/auth-realm: 'Authentication Required'
```

Create basic auth secret:

```bash
htpasswd -c auth username
kubectl create secret generic basic-auth --from-file=auth
```

#### OAuth2 Proxy

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/auth-url: "https://oauth2.example.com/oauth2/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://oauth2.example.com/oauth2/start?rd=$scheme://$host$request_uri"
```

### Resource Limits

For production deployments, adjust based on load:

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

## Upgrading

```bash
helm upgrade perplexity-mcp ./deploy/chart -f my-values.yaml
```

## Uninstalling

```bash
helm uninstall perplexity-mcp
```

## Monitoring

The deployment includes:
- Liveness probe on `/healthz` (checks every 30s)
- Readiness probe on `/healthz` (checks every 10s)

Integrate with Prometheus:

```yaml
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

## Troubleshooting

### Check pod status

```bash
kubectl get pods -l app.kubernetes.io/name=perplexity-mcp
```

### View logs

```bash
kubectl logs -l app.kubernetes.io/name=perplexity-mcp
```

### Test connectivity

```bash
kubectl port-forward svc/perplexity-mcp 8080:8080
curl http://localhost:8080/healthz
```

### Validate secret

```bash
kubectl get secret perplexity-mcp -o jsonpath='{.data.api-key}' | base64 -d
```

## Production Checklist

- [ ] Use `existingSecret` instead of `perplexity.apiKey`
- [ ] Enable ingress with TLS
- [ ] Configure authentication (basic auth or OAuth)
- [ ] Set appropriate resource limits
- [ ] Enable autoscaling for high availability
- [ ] Configure network policies
- [ ] Set up monitoring and alerts
- [ ] Use dedicated namespace
- [ ] Configure pod security policies
- [ ] Set up backup for secrets
