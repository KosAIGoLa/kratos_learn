package server

import (
	stdhttp "net/http"
	_ "net/http/pprof"

	v1 "admin/api/admin/v1"
	"admin/internal/conf"
	"admin/internal/pkg/jwt"
	"admin/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, admin *service.AdminService, jwtManager *jwt.JWTManager, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			JWTAuth(jwtManager),
		),
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
	v1.RegisterAdminHTTPServer(srv, admin)
	openAPIHandler := openapiv2.NewHandler()
	srv.HandlePrefix("/q/", openAPIHandler)
	// pprof endpoints
	srv.HandlePrefix("/debug/pprof/", stdhttp.DefaultServeMux)
	return srv
}
