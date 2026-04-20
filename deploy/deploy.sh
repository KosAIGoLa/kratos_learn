#!/bin/bash

set -e

NAMESPACE="mining-platform"

echo "=== 微服務平台 Kubernetes 部署腳本 ==="
echo ""

# 檢查 kubectl
if ! command -v kubectl &> /dev/null; then
    echo "錯誤: kubectl 未安裝"
    exit 1
fi

# 檢查集群連接
if ! kubectl cluster-info &> /dev/null; then
    echo "錯誤: 無法連接到 Kubernetes 集群"
    exit 1
fi

echo "✓ Kubernetes 集群已連接"
echo ""

# 創建命名空間
echo "1. 創建命名空間..."
kubectl apply -f 00-namespace.yaml
echo "✓ 命名空間創建完成"
echo ""

# 創建 ConfigMap
echo "2. 創建配置..."
kubectl apply -f 01-configmap.yaml
echo "✓ ConfigMap 創建完成"
echo ""

# 部署基礎設施
echo "3. 部署基礎設施..."
kubectl apply -f 10-infrastructure.yaml
echo "✓ 基礎設施部署完成"
echo ""

# 等待基礎設施就緒
echo "4. 等待基礎設施就緒..."
echo "   等待 MySQL..."
kubectl wait --for=condition=ready pod -l app=mysql -n $NAMESPACE --timeout=120s

echo "   等待 Redis..."
kubectl wait --for=condition=ready pod -l app=redis -n $NAMESPACE --timeout=60s

echo "   等待 MongoDB..."
kubectl wait --for=condition=ready pod -l app=mongodb -n $NAMESPACE --timeout=60s

echo "   等待 RabbitMQ..."
kubectl wait --for=condition=ready pod -l app=rabbitmq -n $NAMESPACE --timeout=60s

echo "   等待 etcd..."
kubectl wait --for=condition=ready pod -l app=etcd -n $NAMESPACE --timeout=60s

echo "✓ 基礎設施就緒"
echo ""

# 部署微服務
echo "5. 部署微服務..."
for file in 20-*.yaml 21-*.yaml 22-*.yaml 23-*.yaml 24-*.yaml 25-*.yaml 26-*.yaml 27-*.yaml 28-*.yaml 29-*.yaml 30-*.yaml; do
    if [ -f "$file" ]; then
        echo "   部署 $file..."
        kubectl apply -f "$file"
    fi
done
echo "✓ 微服務部署完成"
echo ""

# 顯示部署狀態
echo "6. 部署狀態:"
echo ""
echo "命名空間:"
kubectl get namespace $NAMESPACE
echo ""
echo "Pod 狀態:"
kubectl get pods -n $NAMESPACE
echo ""
echo "服務:"
kubectl get services -n $NAMESPACE
echo ""

echo "=== 部署完成 ==="
echo ""
echo "查看日誌:"
echo "  kubectl logs -n $NAMESPACE deployment/<service-name>"
echo ""
echo "端口轉發 (Jaeger UI):"
echo "  kubectl port-forward -n $NAMESPACE svc/jaeger 16686:16686"
echo ""
echo "Swagger UI 地址:"
echo "  http://<node-ip>:<node-port>/q/"
echo ""
