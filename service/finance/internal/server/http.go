package server

import (
	stdhttp "net/http"
	_ "net/http/pprof"

	v1 "finance/api/finance/v1"
	"finance/internal/conf"
	"finance/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, finance *service.FinanceService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
		),
		http.ResponseEncoder(ResponseEncoder()),
		http.ErrorEncoder(ErrorEncoder()),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterFinanceHTTPServer(srv, finance)
	openAPIHandler := openapiv2.NewHandler()
	srv.HandlePrefix("/q/", openAPIHandler)
	// pprof endpoints
	srv.HandlePrefix("/debug/pprof/", stdhttp.DefaultServeMux)
	return srv
}
