package main

import (
	"flag"
	"os"

	"sms/internal/conf"
	"sms/internal/pkg/tracing"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	kratostracing "github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, rr registry.Registrar) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Registrar(rr),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", kratostracing.TraceID(),
		"span.id", kratostracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer func() {
		if err := c.Close(); err != nil {
			_, _ = os.Stderr.WriteString("failed to close config: " + err.Error() + "\n")
		}
	}()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// 初始化 OpenTelemetry 鏈路追蹤
	if bc.Trace != nil && bc.Trace.Endpoint != "" {
		tracerCleanup, err := tracing.InitTracer(bc.Trace.Endpoint, Name, float64(bc.Trace.SampleRate))
		if err != nil {
			if logErr := logger.Log(log.LevelWarn, "msg", "failed to init tracer", "err", err); logErr != nil {
				_, _ = os.Stderr.WriteString("failed to write warn log: " + logErr.Error() + "\n")
			}
		} else {
			defer tracerCleanup()
			if logErr := logger.Log(log.LevelInfo, "msg", "tracer initialized", "endpoint", bc.Trace.Endpoint); logErr != nil {
				_, _ = os.Stderr.WriteString("failed to write info log: " + logErr.Error() + "\n")
			}
		}
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Sms, bc.Registry, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
