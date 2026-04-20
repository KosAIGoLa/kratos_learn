#!/bin/bash

# Script to add etcd service discovery to all services
# Usage: ./scripts/add_etcd_to_services.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Services to update (excluding admin and user which are already done)
SERVICES=("content" "cron" "finance" "order" "payment" "product" "system")

# Base directory
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${GREEN}Starting etcd integration for remaining services...${NC}\n"

for SERVICE in "${SERVICES[@]}"; do
    echo -e "${YELLOW}Processing service: ${SERVICE}${NC}"
    
    SERVICE_DIR="${BASE_DIR}/service/${SERVICE}"
    
    if [ ! -d "$SERVICE_DIR" ]; then
        echo -e "${RED}Service directory not found: ${SERVICE_DIR}${NC}"
        continue
    fi
    
    cd "$SERVICE_DIR"
    
    # 1. Update go.mod
    echo "  - Updating go.mod..."
    if ! grep -q "github.com/go-kratos/kratos/contrib/registry/etcd/v2" go.mod; then
        # Add dependencies to require block
        sed -i.bak '/^require (/a\
	github.com/go-kratos/kratos/contrib/registry/etcd/v2 v2.0.0-20260404020628-f149714c1d54\
	go.etcd.io/etcd/client/v3 v3.5.11
' go.mod
        rm go.mod.bak
    fi
    
    # 2. Update config.yaml
    echo "  - Updating config.yaml..."
    if ! grep -q "registry:" configs/config.yaml; then
        cat >> configs/config.yaml << 'EOF'
registry:
  etcd:
    endpoints:
      - 127.0.0.1:2379
    dial_timeout: 5s
EOF
    fi
    
    # 3. Update conf.proto
    echo "  - Updating conf.proto..."
    CONF_PROTO="internal/conf/conf.proto"
    
    # Add Registry to Bootstrap if not exists
    if ! grep -q "Registry registry" "$CONF_PROTO"; then
        # Find the last field number in Bootstrap
        LAST_FIELD=$(grep -E "^\s+\w+\s+\w+\s+=\s+[0-9]+;" "$CONF_PROTO" | grep -A 5 "message Bootstrap" | tail -1 | sed 's/.*= \([0-9]*\);/\1/')
        NEW_FIELD=$((LAST_FIELD + 1))
        
        # Add Registry field to Bootstrap
        sed -i.bak "/message Bootstrap {/,/^}/ s/^}/  Registry registry = ${NEW_FIELD};\n}/" "$CONF_PROTO"
        
        # Add Registry message definition at the end
        cat >> "$CONF_PROTO" << 'EOF'

message Registry {
  message Etcd {
    repeated string endpoints = 1;
    google.protobuf.Duration dial_timeout = 2;
  }
  Etcd etcd = 1;
}
EOF
        rm "${CONF_PROTO}.bak"
    fi
    
    # 4. Update server.go
    echo "  - Updating server.go..."
    SERVER_GO="internal/server/server.go"
    
    if ! grep -q "NewRegistrar" "$SERVER_GO"; then
        cat > "$SERVER_GO" << EOF
package server

import (
	"${SERVICE}/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar)

// NewRegistrar creates a new etcd registrar
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	if conf == nil || conf.Etcd == nil {
		return nil
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Etcd.Endpoints,
		DialTimeout: conf.Etcd.DialTimeout.AsDuration(),
	})
	if err != nil {
		panic(err)
	}
	return etcd.New(client)
}
EOF
    fi
    
    # 5. Update wire.go
    echo "  - Updating wire.go..."
    WIRE_GO="cmd/${SERVICE}/wire.go"
    
    if [ -f "$WIRE_GO" ]; then
        # Update wireApp function signature
        sed -i.bak 's/func wireApp(\*conf.Server, \*conf.Data,/func wireApp(*conf.Server, *conf.Data,/' "$WIRE_GO"
        sed -i.bak 's/log.Logger) (\*kratos.App,/*conf.Registry, log.Logger) (*kratos.App,/' "$WIRE_GO"
        rm "${WIRE_GO}.bak" 2>/dev/null || true
    fi
    
    # 6. Update main.go
    echo "  - Updating main.go..."
    MAIN_GO="cmd/${SERVICE}/main.go"
    
    if [ -f "$MAIN_GO" ]; then
        # Add registry import if not exists
        if ! grep -q '"github.com/go-kratos/kratos/v2/registry"' "$MAIN_GO"; then
            sed -i.bak '/"github.com\/go-kratos\/kratos\/v2\/transport\/grpc"/a\
	"github.com/go-kratos/kratos/v2/registry"
' "$MAIN_GO"
        fi
        
        # Update newApp function
        sed -i.bak 's/func newApp(logger log.Logger, gs \*grpc.Server, hs \*http.Server)/func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, r registry.Registrar)/' "$MAIN_GO"
        
        # Add Registrar to kratos.New if not exists
        if ! grep -q "kratos.Registrar(r)" "$MAIN_GO"; then
            sed -i.bak '/kratos.Server(/,/),$/ s/),$/),\n\t\tkratos.Registrar(r),/' "$MAIN_GO"
        fi
        
        # Update wireApp call
        sed -i.bak 's/wireApp(bc.Server, bc.Data,/wireApp(bc.Server, bc.Data, bc.Registry,/' "$MAIN_GO"
        
        rm "${MAIN_GO}.bak" 2>/dev/null || true
    fi
    
    echo -e "${GREEN}  ✓ ${SERVICE} updated successfully${NC}\n"
done

echo -e "${GREEN}All services updated!${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo "1. For each service, run:"
echo "   cd service/<service>"
echo "   go mod tidy"
echo "   make api"
echo "   go generate ./..."
echo ""
echo "2. Start etcd:"
echo "   docker run -d --name etcd -p 2379:2379 -p 2380:2380 quay.io/coreos/etcd:latest"
echo ""
echo "3. Start services and verify registration:"
echo "   etcdctl get --prefix /microservices/"
