# Kubernetes Deployment Upgrade Guide for v1.4.0

**English** | [简体中文](./UPGRADE.zh-CN.md)

## 🚨 Important Changes in v1.4.0

### Critical Bug Fix
v1.4.0 fixes a **critical bug** in UnionPay API rate interpretation. **All existing exchange rate data is incorrect** and must be cleaned and re-fetched.

### Deployment Changes

1. **Unified Docker Image**
   - Previous: Separate `rateflow-api` and `rateflow-worker` images
   - **New**: Single `rateflow-api` image contains both API server and worker binary
   - Worker binary path: `/app/rateflow-worker`

2. **New Worker Command: fetch-matrix**
   - Efficiently fetches all combinations of specified currencies
   - Replaces individual currency pair CronJobs
   - Example: 5 currencies = 20 currency pairs in one job

3. **New Worker Command: clean**
   - Safely delete incorrect exchange rate data
   - Supports dry-run mode for safety
   - Filter by currency pair and date range

## 📋 Upgrade Steps

### Step 1: Update Image References

**Before** (v1.3.1):
```yaml
image: tyokyo320/rateflow-worker:latest
args: ["fetch", "--pair", "CNY/JPY"]
```

**After** (v1.4.0):
```yaml
image: tyokyo320/rateflow-api:v1.4.0
command: ["/app/rateflow-worker"]
args: ["fetch", "--pair", "CNY/JPY"]
```

### Step 2: Deploy Updated CronJobs

#### Option A: Use fetch-matrix (Recommended)

Deploy the new batch fetching CronJob:

```bash
# Apply the new fetch-matrix CronJob
kubectl apply -f deploy/k8s/worker/worker-cronjob-matrix.yaml

# Delete old individual currency pair jobs
kubectl delete cronjob rateflow-fetch-cny-jpy -n rateflow
kubectl delete cronjob rateflow-fetch-jpy-usd -n rateflow
```

The new `worker-cronjob-matrix.yaml` fetches multiple currency pairs in a single job:
- CNY/JPY, CNY/USD, CNY/EUR, CNY/GBP
- JPY/CNY, JPY/USD, JPY/EUR, JPY/GBP
- USD/CNY, USD/JPY, USD/EUR, USD/GBP
- EUR/CNY, EUR/JPY, EUR/USD, EUR/GBP
- GBP/CNY, GBP/JPY, GBP/USD, GBP/EUR

**Configure currencies**:
```yaml
args:
  - "fetch-matrix"
  - "--currencies"
  - "CNY,JPY,USD,EUR,GBP"  # Edit this line
  - "--provider"
  - "unionpay"
```

#### Option B: Update Existing Individual Jobs

If you prefer to keep individual CronJobs:

```bash
# Update existing CronJobs with new image
kubectl set image cronjob/rateflow-fetch-cny-jpy worker=tyokyo320/rateflow-api:v1.4.0 -n rateflow
kubectl set image cronjob/rateflow-fetch-jpy-usd worker=tyokyo320/rateflow-api:v1.4.0 -n rateflow

# Patch to add command
kubectl patch cronjob rateflow-fetch-cny-jpy -n rateflow --type='json' \
  -p='[{"op": "add", "path": "/spec/jobTemplate/spec/template/spec/containers/0/command", "value": ["/app/rateflow-worker"]}]'
```

### Step 3: Clean Old Data

Run a Kubernetes Job to clean incorrect data:

```bash
# Create a one-time Job to clean all data
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
        args: ["clean", "--dry-run"]  # Remove --dry-run to actually delete
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

# Check the logs to see what will be deleted
kubectl logs job/rateflow-clean-all -n rateflow

# If dry-run looks good, run actual deletion
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
        args: ["clean"]  # Actual deletion - will prompt for confirmation
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

# Connect to the pod to confirm deletion
kubectl attach -it job/rateflow-clean-all -n rateflow
# Type 'yes' when prompted
```

**Alternative: Clean via database directly**:
```bash
# Connect to PostgreSQL pod
kubectl exec -it statefulset/rateflow-postgres -n rateflow -- psql -U rateflow -d rateflow

# In psql:
TRUNCATE TABLE exchange_rates;
\q
```

### Step 4: Fetch Fresh Data

Manually trigger a fetch to populate with correct data:

```bash
# Trigger the fetch-matrix CronJob manually
kubectl create job --from=cronjob/rateflow-fetch-matrix rateflow-fetch-matrix-manual -n rateflow

# Monitor the job
kubectl logs job/rateflow-fetch-matrix-manual -n rateflow -f

# Or use a one-time Job
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
          - "2024-11-01"  # Fetch historical data
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

### Step 5: Verify Data

```bash
# Check API health
kubectl port-forward svc/rateflow-api 8080:8080 -n rateflow
curl http://localhost:8080/health

# Verify currency pairs work (especially USD/JPY which was broken before)
curl "http://localhost:8080/api/v1/rates/latest?pair=USD/JPY"
curl "http://localhost:8080/api/v1/rates/latest?pair=JPY/USD"
curl "http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY"

# Check database directly
kubectl exec -it statefulset/rateflow-postgres -n rateflow -- \
  psql -U rateflow -d rateflow -c \
  "SELECT base_currency, quote_currency, value, effective_date
   FROM exchange_rates
   ORDER BY effective_date DESC, base_currency, quote_currency
   LIMIT 20;"
```

**Expected results** (approximate, November 2024):
```
base_currency | quote_currency | value   | effective_date
--------------+----------------+---------+---------------
CNY           | JPY            | 21.34   | 2024-11-08
JPY           | USD            | 0.0065  | 2024-11-08   ← Previously returned 404!
USD           | JPY            | 153.55  | 2024-11-08   ← Previously returned 404!
USD           | CNY            | 7.17    | 2024-11-08
```

## 📦 Kustomize Deployment

If using Kustomize:

```bash
# Update the deployment
kubectl apply -k deploy/k8s/

# Or specific components
kubectl apply -k deploy/k8s/worker/
kubectl apply -k deploy/k8s/api/
```

The `deploy/k8s/worker/kustomization.yaml` has been updated to use `worker-cronjob-matrix.yaml` by default.

## 🔧 Configuration

### Image Tag Update

Update `deploy/k8s/api/kustomization.yaml`:
```yaml
images:
  - name: tyokyo320/rateflow-api
    newTag: v1.4.0  # Update from latest or v1.3.1
```

Update `deploy/k8s/worker/kustomization.yaml`:
```yaml
images:
  - name: tyokyo320/rateflow-api  # Changed from rateflow-worker
    newTag: v1.4.0
```

## ⚠️ Troubleshooting

### CronJob Not Running

```bash
# Check CronJob status
kubectl get cronjobs -n rateflow

# Check recent jobs
kubectl get jobs -n rateflow --sort-by=.metadata.creationTimestamp

# Check pod logs
kubectl logs -l app=rateflow-worker -n rateflow --tail=100
```

### Data Still Incorrect

1. Verify you're using v1.4.0:
```bash
kubectl describe cronjob rateflow-fetch-matrix -n rateflow | grep Image
```

2. Check logs for "inverted" keyword (indicates new logic):
```bash
kubectl logs -l app=rateflow-worker -n rateflow | grep inverted
```

3. Ensure old data was cleaned:
```bash
kubectl exec -it statefulset/rateflow-postgres -n rateflow -- \
  psql -U rateflow -d rateflow -c \
  "SELECT created_at, COUNT(*) FROM exchange_rates GROUP BY created_at ORDER BY created_at DESC;"
```

### Worker Binary Not Found

Error: `exec: "/app/rateflow-worker": stat /app/rateflow-worker: no such file or directory`

**Cause**: Using old image that doesn't include worker binary.

**Solution**: Ensure you're using v1.4.0+ image:
```bash
kubectl set image cronjob/rateflow-fetch-matrix worker=tyokyo320/rateflow-api:v1.4.0 -n rateflow
```

## 📚 Additional Resources

- [Main Migration Guide](../../MIGRATION_GUIDE.md)
- [中文迁移指南](../../docs/MIGRATION.zh-CN.md)
- [Kubernetes Deployment README](./README.md)
- [Kubernetes部署文档](./README_CN.md)

## ✅ Upgrade Checklist

- [ ] Update API deployment image to v1.4.0
- [ ] Update or replace worker CronJobs
- [ ] Verify new image includes `/app/rateflow-worker` binary
- [ ] Clean old incorrect exchange rate data
- [ ] Fetch fresh data using fetch-matrix or individual fetch commands
- [ ] Verify all currency pairs work (especially USD/JPY, JPY/USD)
- [ ] Update monitoring/alerting for new CronJob names
- [ ] Update documentation if you have custom deployment processes

---

**Upgrade successfully!** 🚀

For questions or issues, please open an issue at https://github.com/tyokyo320/rateflow/issues
