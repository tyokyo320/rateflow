[English](README.md) | [中文](README_CN.md)

# Kubernetes 部署指南

本目录包含 RateFlow 的 Kubernetes 部署配置，使用 Kustomize 进行管理。

## 目录结构

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

## 快速开始

### 1. 修改配置

**必须修改的配置：**

1. `base/secret.yaml` - 修改数据库密码
2. `base/ingress.yaml` - 修改域名和 Ingress Controller
3. `api/kustomization.yaml` - 修改 API 镜像标签（如 `v1.0.0`）
4. `worker/kustomization.yaml` - 修改 worker 镜像标签（如 `v1.0.0`）

**可选修改：**

1. `base/configmap.yaml` - 根据需要调整配置
2. `postgres/pvc.yaml` - 调整存储类和容量
3. `postgres/statefulset.yaml` - 调整资源限制
4. `api/api-deployment.yaml` - 调整副本数和资源限制

### 2. 部署到集群

使用 Kustomize 部署（推荐）：

```bash
# 预览将要部署的资源
kubectl kustomize deploy/k8s

# 部署到集群
kubectl apply -k deploy/k8s

# 查看部署状态
kubectl get all -n rateflow
```

或者按模块部署：

```bash
# 按顺序部署各个模块
kubectl apply -k deploy/k8s/base        # Base resources (namespace, configmap, secret, ingress)
kubectl apply -k deploy/k8s/postgres    # PostgreSQL
kubectl apply -k deploy/k8s/redis       # Redis
kubectl apply -k deploy/k8s/api         # API service
kubectl apply -k deploy/k8s/worker      # Worker cronjobs
```

### 3. 初始化数据库

数据库会自动创建表结构（GORM AutoMigrate），但你需要手动获取初始汇率数据：

```bash
# 手动触发一次 worker 任务
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

# 或者等待 CronJob 自动执行（每小时）
kubectl get cronjobs -n rateflow
```

### 4. 访问服务

```bash
# 查看 Ingress 地址
kubectl get ingress -n rateflow

# 端口转发（本地测试）
kubectl port-forward -n rateflow svc/rateflow-api 8080:8080

# 访问 http://localhost:8080
```

## 常用命令

### 查看日志

```bash
# API 日志
kubectl logs -n rateflow -l app=rateflow-api --tail=100 -f

# Worker 日志 (CNY/JPY)
kubectl logs -n rateflow -l job-name=rateflow-fetch-cny-jpy --tail=100

# Worker 日志 (JPY/USD)
kubectl logs -n rateflow -l job-name=rateflow-fetch-jpy-usd --tail=100

# PostgreSQL 日志
kubectl logs -n rateflow -l app=postgres --tail=100 -f
```

### 调试

```bash
# 进入 API Pod
kubectl exec -it -n rateflow deployment/rateflow-api -- /bin/sh

# 进入 PostgreSQL Pod
kubectl exec -it -n rateflow statefulset/postgres -- psql -U rateflow -d rateflow

# 查看环境变量
kubectl exec -n rateflow deployment/rateflow-api -- env

# 检查健康状态
kubectl exec -n rateflow deployment/rateflow-api -- wget -qO- http://localhost:8080/health
```

### 扩缩容

```bash
# 扩展 API 实例
kubectl scale deployment rateflow-api -n rateflow --replicas=3

# 查看 HPA（如果配置了）
kubectl get hpa -n rateflow
```

### 更新部署

```bash
# 更新镜像版本（在 kustomization.yaml 中修改）
# 编辑 api/kustomization.yaml 和 worker/kustomization.yaml 中的 newTag
kubectl apply -k deploy/k8s

# 或者直接修改镜像
kubectl set image deployment/rateflow-api -n rateflow \
  api=tyokyo320/rateflow-api:v1.0.1

# 查看滚动更新状态
kubectl rollout status deployment/rateflow-api -n rateflow

# 回滚
kubectl rollout undo deployment/rateflow-api -n rateflow
```

### 清理

```bash
# 删除所有资源
kubectl delete -k deploy/k8s

# 或者删除命名空间（会删除所有资源）
kubectl delete namespace rateflow
```

## 添加新的货币对

要追踪更多货币对（如 EUR/USD, GBP/USD），创建新的 CronJob 文件：

### 示例：添加 EUR/USD

1. 创建新的 CronJob 文件：

```bash
# 创建 deploy/k8s/worker/worker-cronjob-eur-usd.yaml
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
  schedule: "10 * * * *"  # 每小时第 10 分钟执行
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

1. 添加到 `worker/kustomization.yaml`：

```yaml
resources:
  - worker-cronjob.yaml
  - worker-cronjob-jpy-usd.yaml
  - worker-cronjob-eur-usd.yaml  # 添加此行
```

1. 部署：

```bash
kubectl apply -k deploy/k8s/worker
```

**提示：**

- 使用不同的时间偏移（0, 5, 10, 15 分钟）避免所有任务同时运行
- 根据实际使用情况调整资源限制
- 监控任务历史：`kubectl get jobs -n rateflow`

## 镜像版本管理

### 标签策略

生产环境建议使用语义化版本：

```bash
# 构建并标记镜像
docker build -t tyokyo320/rateflow-api:v1.0.0 -f Dockerfile .
docker build -t tyokyo320/rateflow-worker:v1.0.0 -f deploy/docker/worker.Dockerfile .

# 推送到镜像仓库
docker push tyokyo320/rateflow-api:v1.0.0
docker push tyokyo320/rateflow-worker:v1.0.0

# 同时标记为 latest（可选）
docker tag tyokyo320/rateflow-api:v1.0.0 tyokyo320/rateflow-api:latest
docker tag tyokyo320/rateflow-worker:v1.0.0 tyokyo320/rateflow-worker:latest
docker push tyokyo320/rateflow-api:latest
docker push tyokyo320/rateflow-worker:latest
```

### 在 Kustomize 中更新镜像标签

编辑 `api/kustomization.yaml`：

```yaml
images:
  - name: tyokyo320/rateflow-api
    newTag: v1.0.1  # 新版本
```

编辑 `worker/kustomization.yaml`：

```yaml
images:
  - name: tyokyo320/rateflow-worker
    newTag: v1.0.1  # 新版本
```

然后部署：

```bash
kubectl apply -k deploy/k8s
```

## 生产环境建议

### 1. 持久化存储

修改 `postgres/pvc.yaml` 中的 `storageClassName`：

```yaml
spec:
  storageClassName: your-storage-class  # 例如: gp2, standard, nfs-client
  resources:
    requests:
      storage: 20Gi  # 根据需求调整
```

### 2. TLS/HTTPS

在 `base/ingress.yaml` 中启用 TLS：

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

需要先安装 cert-manager：

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

### 3. 资源限制

根据实际负载调整资源：

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### 4. 监控和告警

添加 Prometheus 注解：

```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
```

### 5. 密钥管理

生产环境建议使用外部密钥管理：

- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
- [External Secrets Operator](https://external-secrets.io/)
- [Vault](https://www.vaultproject.io/)

### 6. 高可用性

- PostgreSQL: 使用 [CloudNativePG](https://cloudnative-pg.io/) 或托管服务
- Redis: 使用 Redis Sentinel 或 Redis Cluster
- API: 至少 2 个副本 + Pod Anti-Affinity

## 故障排查

### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl get pods -n rateflow

# 查看详细信息
kubectl describe pod <pod-name> -n rateflow

# 查看日志
kubectl logs <pod-name> -n rateflow
```

### 数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
kubectl get pods -n rateflow -l app=postgres

# 测试数据库连接
kubectl run -it --rm psql-test \
  --image=postgres:17-alpine \
  --restart=Never \
  --namespace=rateflow \
  -- psql -h postgres -U rateflow -d rateflow
```

### Ingress 无法访问

```bash
# 检查 Ingress
kubectl get ingress -n rateflow
kubectl describe ingress rateflow-ingress -n rateflow

# 检查 Ingress Controller
kubectl get pods -n ingress-nginx
```

### CronJob 未运行

```bash
# 检查 CronJob 状态
kubectl get cronjobs -n rateflow

# 查看最近的 Job
kubectl get jobs -n rateflow

# 检查特定 CronJob
kubectl describe cronjob rateflow-fetch-cny-jpy -n rateflow

# 手动触发一个 Job 测试
kubectl create job --from=cronjob/rateflow-fetch-cny-jpy manual-test-1 -n rateflow
```

## 参考资料

- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [Kustomize 文档](https://kustomize.io/)
- [kubectl 速查表](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
