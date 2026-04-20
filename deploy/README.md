# Kubernetes 部署

本目錄包含微服務平台的 Kubernetes 部署配置。

## 部署文件說明

| 文件 | 說明 |
|------|------|
| `00-namespace.yaml` | 創建命名空間 `mining-platform` |
| `01-configmap.yaml` | 所有服務的配置文件 |
| `10-infrastructure.yaml` | 基礎設施：MySQL、Redis、MongoDB、RabbitMQ、etcd、Jaeger |
| `20-content-service.yaml` | Content 服務部署 |
| `21-finance-service.yaml` | Finance 服務部署 |
| `22-order-service.yaml` | Order 服務部署 |
| `23-payment-service.yaml` | Payment 服務部署 |
| `24-product-service.yaml` | Product 服務部署 |
| `25-system-service.yaml` | System 服務部署 |
| `26-user-service.yaml` | User 服務部署 |
| `27-admin-service.yaml` | Admin 服務部署 |
| `28-mail-service.yaml` | Mail 服務部署 |
| `29-sms-service.yaml` | SMS 服務部署 |
| `30-cron-service.yaml` | Cron 服務部署 |

## 快速部署

### 1. 部署全部服務

```bash
# 部署所有資源
kubectl apply -f .

# 或按順序部署
kubectl apply -f 00-namespace.yaml
kubectl apply -f 01-configmap.yaml
kubectl apply -f 10-infrastructure.yaml
kubectl apply -f 20-content-service.yaml
kubectl apply -f 21-finance-service.yaml
kubectl apply -f 22-order-service.yaml
kubectl apply -f 23-payment-service.yaml
kubectl apply -f 24-product-service.yaml
kubectl apply -f 25-system-service.yaml
kubectl apply -f 26-user-service.yaml
kubectl apply -f 27-admin-service.yaml
kubectl apply -f 28-mail-service.yaml
kubectl apply -f 29-sms-service.yaml
kubectl apply -f 30-cron-service.yaml
```

### 2. 查看部署狀態

```bash
# 查看所有 Pod
kubectl get pods -n mining-platform

# 查看服務
kubectl get services -n mining-platform

# 查看基礎設施日誌
kubectl logs -n mining-platform deployment/mysql
kubectl logs -n mining-platform deployment/redis
kubectl logs -n mining-platform deployment/etcd

# 查看服務日誌
kubectl logs -n mining-platform deployment/user-service
kubectl logs -n mining-platform deployment/admin-service
```

### 3. 訪問服務

**端口轉發（本地開發）：**
```bash
# MySQL
kubectl port-forward -n mining-platform svc/mysql 3306:3306

# Redis
kubectl port-forward -n mining-platform svc/redis 6379:6379

# Jaeger UI
kubectl port-forward -n mining-platform svc/jaeger 16686:16686

# User Service HTTP
kubectl port-forward -n mining-platform svc/user-service 8107:8107
```

**使用 Ingress（生產環境）：**
- Content: http://content.local
- Finance: http://finance.local
- Order: http://order.local
- Payment: http://payment.local
- Product: http://product.local
- System: http://system.local
- User: http://user.local
- Admin: http://admin.local
- Mail: http://mail.local
- SMS: http://sms.local
- Cron: http://cron.local

## 構建 Docker 映像

在每個服務目錄下創建 `Dockerfile`：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/*/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8101 9101
CMD ["./server"]
```

構建命令：
```bash
# 構建所有服務映像
for service in content finance order payment product system user admin mail sms cron; do
  docker build -t ${service}-service:latest ./service/${service}
done
```

## 清理部署

```bash
# 刪除所有資源
kubectl delete -f .

# 或刪除命名空間（會刪除所有內容）
kubectl delete namespace mining-platform
```

## 配置說明

### 資源限制
- 每個微服務：128Mi 內存 / 100m CPU（請求），256Mi 內存 / 200m CPU（限制）
- MySQL：512Mi 內存 / 250m CPU（請求），1Gi 內存 / 500m CPU（限制）
- Redis：128Mi 內存 / 100m CPU
- MongoDB：256Mi 內存 / 100m CPU
- RabbitMQ：256Mi 內存 / 100m CPU

### 健康檢查
- 所有服務都配置了 livenessProbe 和 readinessProbe
- HTTP 健康檢查路徑：`/q/health`

### 鏈路追蹤
- 所有服務已配置 OpenTelemetry
- Jaeger UI: http://localhost:16686
