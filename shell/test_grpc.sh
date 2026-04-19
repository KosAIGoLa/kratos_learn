#!/bin/bash

# gRPC 服務測試腳本
# 使用 grpcurl 測試所有微服務的 gRPC 接口
#
# 安裝 grpcurl:
#   brew install grpcurl (macOS)
#   或 https://github.com/fullstorydev/grpcurl/releases
#
# 使用方法:
#   ./test_grpc.sh [service_name] [method_name]
#   ./test_grpc.sh all          # 測試所有服務健康狀態
#   ./test_grpc.sh user         # 測試 user 服務
#   ./test_grpc.sh user Login   # 測試 user 服務的 Login 方法

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 服務配置
SERVICES=(
    "content:9101:content.v1.Content"
    "finance:9102:finance.v1.Finance"
    "order:9103:order.v1.Order"
    "payment:9104:payment.v1.Payment"
    "product:9105:product.v1.Product"
    "system:9106:system.v1.System"
    "user:9107:user.v1.User"
    "admin:9108:admin.v1.Admin"
)

# 檢查 grpcurl 是否安裝
if ! command -v grpcurl &> /dev/null; then
    echo -e "${RED}Error: grpcurl 未安裝${NC}"
    echo "請先安裝 grpcurl:"
    echo "  macOS: brew install grpcurl"
    echo "  Linux: https://github.com/fullstorydev/grpcurl/releases"
    exit 1
fi

# 顯示幫助信息
show_help() {
    echo "gRPC 服務測試腳本"
    echo ""
    echo "使用方法:"
    echo "  $0 all                          # 測試所有服務健康狀態"
    echo "  $0 list                         # 列出所有可用服務和方法"
    echo "  $0 <service>                    # 測試指定服務 (帶示例請求)"
    echo "  $0 <service> <method>           # 調用指定方法"
    echo "  $0 <service> <method> '<json>'  # 調用方法並傳入 JSON 參數"
    echo ""
    echo "服務列表:"
    for svc in "${SERVICES[@]}"; do
        IFS=':' read -r name port proto <<< "$svc"
        echo "  - $name (port: $port)"
    done
    echo ""
    echo "示例:"
    echo "  $0 user                         # 測試 user 服務"
    echo "  $0 user List                    # 查看 user 服務的方法列表"
    echo "  $0 user Login '{\"phone\": \"13800138000\", \"password\": \"123456\"}'"
}

# 獲取服務信息
get_service_info() {
    local name=$1
    for svc in "${SERVICES[@]}"; do
        IFS=':' read -r svc_name port proto <<< "$svc"
        if [[ "$svc_name" == "$name" ]]; then
            echo "$port:$proto"
            return 0
        fi
    done
    return 1
}

# 列出所有服務
list_services() {
    echo -e "${BLUE}=== 服務列表 ===${NC}"
    for svc in "${SERVICES[@]}"; do
        IFS=':' read -r name port proto <<< "$svc"
        echo -e "${GREEN}$name${NC} - localhost:$port ($proto)"
        # 嘗試獲取方法列表
        echo "  可用方法:"
        grpcurl -plaintext "localhost:$port" list "$proto" 2>/dev/null | sed 's/^/    /' || echo "    (無法連接)"
        echo ""
    done
}

# 測試服務健康狀態
check_health() {
    local name=$1
    local port=$2
    local proto=$3
    
    echo -e "${BLUE}Testing $name (localhost:$port)...${NC}"
    
    # 使用 grpcurl 檢查服務是否可達
    if grpcurl -plaintext "localhost:$port" list "$proto" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓ 服務正常${NC}"
        # 顯示可用方法
        echo "  可用方法:"
        grpcurl -plaintext "localhost:$port" list "$proto" 2>/dev/null | sed 's/^/    /'
    else
        echo -e "  ${RED}✗ 無法連接${NC}"
    fi
    echo ""
}

# 測試所有服務
test_all() {
    echo -e "${BLUE}=== 測試所有 gRPC 服務 ===${NC}"
    echo ""
    for svc in "${SERVICES[@]}"; do
        IFS=':' read -r name port proto <<< "$svc"
        check_health "$name" "$port" "$proto"
    done
}

# 調用服務方法
call_method() {
    local service=$1
    local method=$2
    local json_data=$3
    
    local info
    info=$(get_service_info "$service")
    if [[ -z "$info" ]]; then
        echo -e "${RED}Error: 未知的服務 '$service'${NC}"
        exit 1
    fi
    
    IFS=':' read -r port proto <<< "$info"
    
    echo -e "${BLUE}=== 調用 $service.$method ===${NC}"
    echo -e "${YELLOW}地址: localhost:$port${NC}"
    echo -e "${YELLOW}方法: $proto/$method${NC}"
    
    if [[ -n "$json_data" ]]; then
        echo -e "${YELLOW}請求: $json_data${NC}"
        echo ""
        grpcurl -plaintext -d "$json_data" "localhost:$port" "$proto/$method"
    else
        echo ""
        grpcurl -plaintext "localhost:$port" "$proto/$method"
    fi
}

# 執行服務示例測試
test_service_examples() {
    local service=$1
    
    local info
    info=$(get_service_info "$service")
    if [[ -z "$info" ]]; then
        echo -e "${RED}Error: 未知的服務 '$service'${NC}"
        exit 1
    fi
    
    IFS=':' read -r port proto <<< "$info"
    
    echo -e "${BLUE}=== 測試 $service 服務 (localhost:$port) ===${NC}"
    echo ""
    
    # 檢查服務是否可達
    if ! grpcurl -plaintext "localhost:$port" list "$proto" > /dev/null 2>&1; then
        echo -e "${RED}✗ 無法連接到 $service 服務${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ 服務可連接${NC}"
    echo ""
    echo -e "${BLUE}可用方法:${NC}"
    grpcurl -plaintext "localhost:$port" list "$proto" | sed 's/^/  /'
    echo ""
    
    # 根據服務執行示例請求
    case $service in
        "user")
            echo -e "${BLUE}--- 示例: 用戶登錄 ---${NC}"
            grpcurl -plaintext -d '{"phone": "13800138000", "password": "123456"}' "localhost:$port" "$proto/Login" || echo "(登錄請求失敗)"
            ;;
        "content")
            echo -e "${BLUE}--- 示例: 獲取橫幅列表 ---${NC}"
            grpcurl -plaintext -d '{"type": "home", "status": 1}' "localhost:$port" "$proto/ListBanners" || echo "(請求失敗)"
            ;;
        "finance")
            echo -e "${BLUE}--- 示例: 獲取用戶餘額 ---${NC}"
            grpcurl -plaintext -d '{"user_id": 1}' "localhost:$port" "$proto/GetUserBalance" || echo "(請求失敗)"
            ;;
        "order")
            echo -e "${BLUE}--- 示例: 獲取訂單列表 ---${NC}"
            grpcurl -plaintext -d '{"user_id": 1, "page": 1, "page_size": 10}' "localhost:$port" "$proto/ListOrders" || echo "(請求失敗)"
            ;;
        "product")
            echo -e "${BLUE}--- 示例: 獲取產品列表 ---${NC}"
            grpcurl -plaintext -d '{"status": 1, "page": 1, "page_size": 10}' "localhost:$port" "$proto/ListProducts" || echo "(請求失敗)"
            ;;
        "payment")
            echo -e "${BLUE}--- 示例: 獲取支付渠道列表 ---${NC}"
            grpcurl -plaintext "localhost:$port" "$proto/ListChannels" || echo "(請求失敗)"
            ;;
        "system")
            echo -e "${BLUE}--- 示例: 獲取系統配置 ---${NC}"
            grpcurl -plaintext "localhost:$port" "$proto/ListConfigs" || echo "(請求失敗)"
            ;;
        "admin")
            echo -e "${BLUE}--- 示例: 管理員登錄 ---${NC}"
            grpcurl -plaintext -d '{"username": "admin", "password": "admin123"}' "localhost:$port" "$proto/Login" || echo "(請求失敗)"
            ;;
    esac
}

# 主邏輯
main() {
    case "${1:-help}" in
        "help"|"-h"|"--help")
            show_help
            ;;
        "all")
            test_all
            ;;
        "list")
            list_services
            ;;
        *)
            service=$1
            method=${2:-}
            json_data=${3:-}
            
            if [[ -z "$method" ]]; then
                test_service_examples "$service"
            else
                call_method "$service" "$method" "$json_data"
            fi
            ;;
    esac
}

main "$@"
