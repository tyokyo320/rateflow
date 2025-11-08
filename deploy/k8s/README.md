[English](README.md) | [中文](README_CN.md)

# Kubernetes Deployment Guide

**🔄 Upgrading from v1.3.1?** See [Upgrade Guide](./UPGRADE.md) | [升级指南](./UPGRADE.zh-CN.md)

This directory contains Kubernetes deployment configurations for RateFlow v1.4.0, managed with Kustomize.

## Directory Structure

```
deploy/k8s/
├── kustomization.yaml       # Main Kustomize configuration
├── base/                    # Base resources
│   ├── namespace.yaml       # Namespace
│   ├── configmap.yaml       # ConfigMap
│   ├── secret.yaml          # Secrets (passwords, etc.)
│   ├── ingress.yaml         # Ingress configuration
│   └── kustomization.yaml   # Base kustomization
├── postgres/                # PostgreSQL database
│   ├── pvc.yaml            # PersistentVolumeClaim
│   ├── service.yaml        # Service
│   ├── statefulset.yaml    # StatefulSet
│   └── kustomization.yaml  # Postgres kustomization
├── redis/                   # Redis cache
│   ├── service.yaml        # Service
│   ├── deployment.yaml     # Deployment
│   └── kustomization.yaml  # Redis kustomization
├── api/                     # API service
│   ├── api-deployment.yaml # Deployment
│   ├── api-service.yaml    # Service
│   └── kustomization.yaml  # API kustomization
└── worker/                  # Worker cronjobs
    ├── worker-cronjob.yaml      # CronJob for CNY/JPY
    ├── worker-cronjob-jpy-usd.yaml  # CronJob for JPY/USD
    └── kustomization.yaml       # Worker kustomization
```

## Quick Start

### 1. Modify Configuration

**Required modifications:**

1. `base/secret.yaml` - Change database password
2. `base/ingress.yaml` - Change domain name and Ingress Controller
3. `api/kustomization.yaml` - Change API image tag (e.g., `v1.0.0`)
4. `worker/kustomization.yaml` - Change worker image tag (e.g., `v1.0.0`)

**Optional modifications:**

1. `base/configmap.yaml` - Adjust configuration as needed
2. `postgres/pvc.yaml` - Adjust storage class and capacity
3. `postgres/statefulset.yaml` - Adjust resource limits
4. `api/api-deployment.yaml` - Adjust replicas and resource limits

### 2. Deploy to Cluster

Deploy using Kustomize (recommended):

```bash
# Preview resources to be deployed
kubectl kustomize deploy/k8s

# Deploy to cluster
kubectl apply -k deploy/k8s

# Check deployment status
kubectl get all -n rateflow
```

Or deploy by module:

```bash
# Deploy modules in order
kubectl apply -k deploy/k8s/base        # Base resources (namespace, configmap, secret, ingress)
kubectl apply -k deploy/k8s/postgres    # PostgreSQL
kubectl apply -k deploy/k8s/redis       # Redis
kubectl apply -k deploy/k8s/api         # API service
kubectl apply -k deploy/k8s/worker      # Worker cronjobs
```

### 3. Initialize Database

The database schema is automatically created (GORM AutoMigrate), but you need to manually fetch initial rate data:

```bash
# Manually trigger a worker task
kubectl run -it --rm rateflow-init \
  --image=tyokyo320/rateflow-worker:v1.0.0 \
  --restart=Never \
  --namespace=rateflow \
  --env="DB_HOST=postgres" \
  --env="DB_PORT=5432" \
  --env="DB_USER=rateflow" \
  --env="DB_NAME=rateflow" \
  --env="DB_PASSWORD=your_password" \
  --env="DB_SSLMODE=disable" \
  -- fetch --pair CNY/JPY

# Or wait for CronJob to run automatically (every hour)
kubectl get cronjobs -n rateflow
```

### 4. Access Service

```bash
# View Ingress address
kubectl get ingress -n rateflow

# Port forwarding (local testing)
kubectl port-forward -n rateflow svc/rateflow-api 8080:8080

# Access http://localhost:8080
```

## Common Commands

### View Logs

```bash
# API logs
kubectl logs -n rateflow -l app=rateflow-api --tail=100 -f

# Worker logs (CNY/JPY)
kubectl logs -n rateflow -l job-name=rateflow-fetch-cny-jpy --tail=100

# Worker logs (JPY/USD)
kubectl logs -n rateflow -l job-name=rateflow-fetch-jpy-usd --tail=100

# PostgreSQL logs
kubectl logs -n rateflow -l app=postgres --tail=100 -f
```

### Debug

```bash
# Enter API Pod
kubectl exec -it -n rateflow deployment/rateflow-api -- /bin/sh

# Enter PostgreSQL Pod
kubectl exec -it -n rateflow statefulset/postgres -- psql -U rateflow -d rateflow

# View environment variables
kubectl exec -n rateflow deployment/rateflow-api -- env

# Check health status
kubectl exec -n rateflow deployment/rateflow-api -- wget -qO- http://localhost:8080/health
```

### Scaling

```bash
# Scale API instances
kubectl scale deployment rateflow-api -n rateflow --replicas=3

# View HPA (if configured)
kubectl get hpa -n rateflow
```

### Update Deployment

```bash
# Update image version (modify in kustomization.yaml)
# Edit newTag in api/kustomization.yaml and worker/kustomization.yaml
kubectl apply -k deploy/k8s

# Or directly modify image
kubectl set image deployment/rateflow-api -n rateflow \
  api=tyokyo320/rateflow-api:v1.0.1

# View rolling update status
kubectl rollout status deployment/rateflow-api -n rateflow

# Rollback
kubectl rollout undo deployment/rateflow-api -n rateflow
```

### Cleanup

```bash
# Delete all resources
kubectl delete -k deploy/k8s

# Or delete namespace (will delete all resources)
kubectl delete namespace rateflow
```

## Adding New Currency Pairs

To track additional currency pairs (e.g., EUR/USD, GBP/USD), create a new CronJob file:

### Example: Adding EUR/USD

1. Create a new CronJob file:

```bash
# Create deploy/k8s/worker/worker-cronjob-eur-usd.yaml
cat > deploy/k8s/worker/worker-cronjob-eur-usd.yaml <<'EOF'
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: rateflow-fetch-eur-usd
  namespace: rateflow
  labels:
    app: rateflow-worker
    component: worker
    currency-pair: EUR-USD
spec:
  schedule: "10 * * * *"  # Run every hour at 10 minutes past
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  concurrencyPolicy: Forbid
  jobTemplate:
    metadata:
      labels:
        app: rateflow-worker
        currency-pair: EUR-USD
    spec:
      backoffLimit: 3
      template:
        spec:
          restartPolicy: OnFailure
          containers:
          - name: worker
            image: tyokyo320/rateflow-worker:latest
            imagePullPolicy: Always
            args: ["fetch", "--pair", "EUR/USD"]
            envFrom:
            - configMapRef:
                name: rateflow-config
            env:
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: rateflow-secret
                  key: DB_PASSWORD
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: rateflow-secret
                  key: REDIS_PASSWORD
            resources:
              requests:
                memory: "64Mi"
                cpu: "100m"
              limits:
                memory: "128Mi"
                cpu: "200m"
EOF
```

1. Add to `worker/kustomization.yaml`:

```yaml
resources:
  - worker-cronjob.yaml
  - worker-cronjob-jpy-usd.yaml
  - worker-cronjob-eur-usd.yaml  # Add this line
```

1. Deploy:

```bash
kubectl apply -k deploy/k8s/worker
```

**Tips:**

- Use different schedule offsets (0, 5, 10, 15 minutes) to avoid all jobs running simultaneously
- Adjust resource limits based on actual usage
- Monitor job history: `kubectl get jobs -n rateflow`

## Image Versioning

### Tagging Strategy

Use semantic versioning for production deployments:

```bash
# Build and tag images
docker build -t tyokyo320/rateflow-api:v1.0.0 -f Dockerfile .
docker build -t tyokyo320/rateflow-worker:v1.0.0 -f deploy/docker/worker.Dockerfile .

# Push to registry
docker push tyokyo320/rateflow-api:v1.0.0
docker push tyokyo320/rateflow-worker:v1.0.0

# Also tag as latest for convenience (optional)
docker tag tyokyo320/rateflow-api:v1.0.0 tyokyo320/rateflow-api:latest
docker tag tyokyo320/rateflow-worker:v1.0.0 tyokyo320/rateflow-worker:latest
docker push tyokyo320/rateflow-api:latest
docker push tyokyo320/rateflow-worker:latest
```

### Update Image Tags in Kustomize

Edit `api/kustomization.yaml`:

```yaml
images:
  - name: tyokyo320/rateflow-api
    newTag: v1.0.1  # New version
```

Edit `worker/kustomization.yaml`:

```yaml
images:
  - name: tyokyo320/rateflow-worker
    newTag: v1.0.1  # New version
```

Then deploy:

```bash
kubectl apply -k deploy/k8s
```

## Production Recommendations

### 1. Persistent Storage

Modify `storageClassName` in `postgres/pvc.yaml`:

```yaml
spec:
  storageClassName: your-storage-class  # e.g., gp2, standard, nfs-client
  resources:
    requests:
      storage: 20Gi  # Adjust based on needs
```

### 2. TLS/HTTPS

Enable TLS in `base/ingress.yaml`:

```yaml
metadata:
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - rate.example.com
    secretName: rateflow-tls
```

Install cert-manager first:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

### 3. Resource Limits

Adjust resources based on actual load:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### 4. Monitoring and Alerting

Add Prometheus annotations:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
```

### 5. Secret Management

For production, use external secret management:

- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
- [External Secrets Operator](https://external-secrets.io/)
- [Vault](https://www.vaultproject.io/)

### 6. High Availability

- PostgreSQL: Use [CloudNativePG](https://cloudnative-pg.io/) or managed services
- Redis: Use Redis Sentinel or Redis Cluster
- API: At least 2 replicas + Pod Anti-Affinity

## Troubleshooting

### Pod Won't Start

```bash
# View Pod status
kubectl get pods -n rateflow

# View detailed information
kubectl describe pod <pod-name> -n rateflow

# View logs
kubectl logs <pod-name> -n rateflow
```

### Database Connection Failed

```bash
# Check if PostgreSQL is running
kubectl get pods -n rateflow -l app=postgres

# Test database connection
kubectl run -it --rm psql-test \
  --image=postgres:17-alpine \
  --restart=Never \
  --namespace=rateflow \
  -- psql -h postgres -U rateflow -d rateflow
```

### Ingress Not Accessible

```bash
# Check Ingress
kubectl get ingress -n rateflow
kubectl describe ingress rateflow-ingress -n rateflow

# Check Ingress Controller
kubectl get pods -n ingress-nginx
```

### CronJob Not Running

```bash
# Check CronJob status
kubectl get cronjobs -n rateflow

# View recent jobs
kubectl get jobs -n rateflow

# Check specific CronJob
kubectl describe cronjob rateflow-fetch-cny-jpy -n rateflow

# Manually trigger a job
kubectl create job --from=cronjob/rateflow-fetch-cny-jpy manual-test-1 -n rateflow
```

## References

- [Kubernetes Official Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
