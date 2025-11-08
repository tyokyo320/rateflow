# Kubernetes 部署升级指南 - v1.4.0

[English](./UPGRADE.md) | **简体中文**

## 🚨 v1.4.0 重要变更

### 关键Bug修复
v1.4.0 修复了银联 API 汇率解析的**严重 bug**。**所有现有的汇率数据都是错误的**,必须清理并重新获取。

### 部署变更

1. **统一的 Docker 镜像**
   - 之前: 分离的 `rateflow-api` 和 `rateflow-worker` 镜像
   - **现在**: 单个 `rateflow-api` 镜像包含 API 服务器和 Worker 二进制文件
   - Worker 二进制文件路径: `/app/rateflow-worker`

2. **新的 Worker 命令: fetch-matrix**
   - 高效地获取指定货币列表的所有组合
   - 替代单个货币对的 CronJob
   - 示例: 5种货币 = 一次任务获取20个货币对

3. **新的 Worker 命令: clean**
   - 安全地删除错误的汇率数据
   - 支持干运行模式以确保安全
   - 可按货币对和日期范围过滤

## 📋 升级步骤

### 第1步: 更新镜像引用

**之前** (v1.3.1):
```yaml
image: tyokyo320/rateflow-worker:latest
args: ["fetch", "--pair", "CNY/JPY"]
```

**之后** (v1.4.0):
```yaml
image: tyokyo320/rateflow-api:v1.4.0
command: ["/app/rateflow-worker"]
args: ["fetch", "--pair", "CNY/JPY"]
```

### 第2步: 部署更新的 CronJob

#### 方案 A: 使用 fetch-matrix (推荐)

部署新的批量获取 CronJob:

```bash
# 应用新的 fetch-matrix CronJob
kubectl apply -f deploy/k8s/worker/worker-cronjob-matrix.yaml

# 删除旧的单个货币对任务
kubectl delete cronjob rateflow-fetch-cny-jpy -n rateflow
kubectl delete cronjob rateflow-fetch-jpy-usd -n rateflow
```

新的 `worker-cronjob-matrix.yaml` 在单个任务中获取多个货币对:
- CNY/JPY, CNY/USD, CNY/EUR, CNY/GBP
- JPY/CNY, JPY/USD, JPY/EUR, JPY/GBP
- USD/CNY, USD/JPY, USD/EUR, USD/GBP
- EUR/CNY, EUR/JPY, EUR/USD, EUR/GBP
- GBP/CNY, GBP/JPY, GBP/USD, GBP/EUR

**配置货币列表**:
```yaml
args:
  - "fetch-matrix"
  - "--currencies"
  - "CNY,JPY,USD,EUR,GBP"  # 编辑此行
  - "--provider"
  - "unionpay"
```

#### 方案 B: 更新现有的单个任务

如果您希望保留单个 CronJob:

```bash
# 使用新镜像更新现有 CronJob
kubectl set image cronjob/rateflow-fetch-cny-jpy worker=tyokyo320/rateflow-api:v1.4.0 -n rateflow
kubectl set image cronjob/rateflow-fetch-jpy-usd worker=tyokyo320/rateflow-api:v1.4.0 -n rateflow

# 修补以添加 command
kubectl patch cronjob rateflow-fetch-cny-jpy -n rateflow --type='json' \
  -p='[{"op": "add", "path": "/spec/jobTemplate/spec/template/spec/containers/0/command", "value": ["/app/rateflow-worker"]}]'
```

### 第3步: 清理旧数据

运行 Kubernetes Job 清理错误数据:

```bash
# 创建一次性 Job 清理所有数据(先干运行)
cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: rateflow-clean-all
  namespace: rateflow
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: worker
        image: tyokyo320/rateflow-api:v1.4.0
        command: ["/app/rateflow-worker"]
        args: ["clean", "--dry-run"]  # 移除 --dry-run 以实际删除
        envFrom:
        - configMapRef:
            name: rateflow-config
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rateflow-secret
              key: DB_PASSWORD
EOF

# 检查日志查看将要删除的内容
kubectl logs job/rateflow-clean-all -n rateflow

# 如果干运行结果正确,执行实际删除
kubectl delete job rateflow-clean-all -n rateflow
cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: rateflow-clean-all
  namespace: rateflow
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: worker
        image: tyokyo320/rateflow-api:v1.4.0
        command: ["/app/rateflow-worker"]
        args: ["clean"]  # 实际删除 - 会提示确认
        envFrom:
        - configMapRef:
            name: rateflow-config
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rateflow-secret
              key: DB_PASSWORD
        stdin: true
        tty: true
EOF

# 连接到 pod 确认删除
kubectl attach -it job/rateflow-clean-all -n rateflow
# 提示时输入 'yes'
```

**替代方案: 直接通过数据库清理**:
```bash
# 连接到 PostgreSQL pod
kubectl exec -it statefulset/rateflow-postgres -n rateflow -- psql -U rateflow -d rateflow

# 在 psql 中:
TRUNCATE TABLE exchange_rates;
\q
```

### 第4步: 获取新数据

手动触发获取以填充正确的数据:

```bash
# 手动触发 fetch-matrix CronJob
kubectl create job --from=cronjob/rateflow-fetch-matrix rateflow-fetch-matrix-manual -n rateflow

# 监控任务
kubectl logs job/rateflow-fetch-matrix-manual -n rateflow -f

# 或使用一次性 Job
cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: rateflow-fetch-initial
  namespace: rateflow
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: worker
        image: tyokyo320/rateflow-api:v1.4.0
        command: ["/app/rateflow-worker"]
        args:
          - "fetch-matrix"
          - "--currencies"
          - "CNY,JPY,USD,EUR,GBP"
          - "--start"
          - "2024-11-01"  # 获取历史数据
          - "--end"
          - "2024-11-08"
        envFrom:
        - configMapRef:
            name: rateflow-config
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rateflow-secret
              key: DB_PASSWORD
EOF
```

### 第5步: 验证数据

```bash
# 检查 API 健康状态
kubectl port-forward svc/rateflow-api 8080:8080 -n rateflow
curl http://localhost:8080/health

# 验证货币对工作(特别是之前失败的 USD/JPY)
curl "http://localhost:8080/api/v1/rates/latest?pair=USD/JPY"
curl "http://localhost:8080/api/v1/rates/latest?pair=JPY/USD"
curl "http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY"

# 直接检查数据库
kubectl exec -it statefulset/rateflow-postgres -n rateflow -- \
  psql -U rateflow -d rateflow -c \
  "SELECT base_currency, quote_currency, value, effective_date
   FROM exchange_rates
   ORDER BY effective_date DESC, base_currency, quote_currency
   LIMIT 20;"
```

**预期结果** (大约值,2024年11月):
```
base_currency | quote_currency | value   | effective_date
--------------+----------------+---------+---------------
CNY           | JPY            | 21.34   | 2024-11-08
JPY           | USD            | 0.0065  | 2024-11-08   ← 之前返回 404!
USD           | JPY            | 153.55  | 2024-11-08   ← 之前返回 404!
USD           | CNY            | 7.17    | 2024-11-08
```

## 📦 Kustomize 部署

如果使用 Kustomize:

```bash
# 更新部署
kubectl apply -k deploy/k8s/

# 或特定组件
kubectl apply -k deploy/k8s/worker/
kubectl apply -k deploy/k8s/api/
```

`deploy/k8s/worker/kustomization.yaml` 已更新为默认使用 `worker-cronjob-matrix.yaml`。

## 🔧 配置

### 镜像标签更新

更新 `deploy/k8s/api/kustomization.yaml`:
```yaml
images:
  - name: tyokyo320/rateflow-api
    newTag: v1.4.0  # 从 latest 或 v1.3.1 更新
```

更新 `deploy/k8s/worker/kustomization.yaml`:
```yaml
images:
  - name: tyokyo320/rateflow-api  # 从 rateflow-worker 更改
    newTag: v1.4.0
```

## ⚠️ 故障排除

### CronJob 未运行

```bash
# 检查 CronJob 状态
kubectl get cronjobs -n rateflow

# 检查最近的任务
kubectl get jobs -n rateflow --sort-by=.metadata.creationTimestamp

# 检查 pod 日志
kubectl logs -l app=rateflow-worker -n rateflow --tail=100
```

### 数据仍然不正确

1. 验证您使用的是 v1.4.0:
```bash
kubectl describe cronjob rateflow-fetch-matrix -n rateflow | grep Image
```

2. 检查日志中的 "inverted" 关键字(表示新逻辑):
```bash
kubectl logs -l app=rateflow-worker -n rateflow | grep inverted
```

3. 确保旧数据已清理:
```bash
kubectl exec -it statefulset/rateflow-postgres -n rateflow -- \
  psql -U rateflow -d rateflow -c \
  "SELECT created_at, COUNT(*) FROM exchange_rates GROUP BY created_at ORDER BY created_at DESC;"
```

### Worker 二进制文件未找到

错误: `exec: "/app/rateflow-worker": stat /app/rateflow-worker: no such file or directory`

**原因**: 使用的旧镜像不包含 worker 二进制文件。

**解决方案**: 确保使用 v1.4.0+ 镜像:
```bash
kubectl set image cronjob/rateflow-fetch-matrix worker=tyokyo320/rateflow-api:v1.4.0 -n rateflow
```

## 📚 其他资源

- [主迁移指南(英文)](../../MIGRATION_GUIDE.md)
- [中文迁移指南](../../docs/MIGRATION.zh-CN.md)
- [Kubernetes 部署文档(英文)](./README.md)
- [Kubernetes部署文档(中文)](./README_CN.md)

## ✅ 升级检查清单

- [ ] 将 API 部署镜像更新到 v1.4.0
- [ ] 更新或替换 worker CronJob
- [ ] 验证新镜像包含 `/app/rateflow-worker` 二进制文件
- [ ] 清理旧的错误汇率数据
- [ ] 使用 fetch-matrix 或单个 fetch 命令获取新数据
- [ ] 验证所有货币对工作(特别是 USD/JPY, JPY/USD)
- [ ] 为新的 CronJob 名称更新监控/告警
- [ ] 如有自定义部署流程,更新文档

---

**升级成功!** 🚀

如有问题,请在 https://github.com/tyokyo320/rateflow/issues 提交 issue
