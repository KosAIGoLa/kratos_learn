package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"cron/internal/conf"
	"cron/internal/cron"
	"cron/internal/pkg/tracing"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	kratostracing "github.com/go-kratos/kratos/v2/middleware/tracing"
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

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, taskManager *cron.TaskManager, scheduler *cron.Scheduler) *kratos.App {
	// 注册定时任务
	if err := taskManager.RegisterTasks(); err != nil {
		log.NewHelper(logger).Errorf("注册定时任务失败: %v", err)
	}

	// 创建 Kratos 应用
	app := kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		// 添加启动钩子来启动 cron 调度器
		kratos.BeforeStart(func(ctx context.Context) error {
			log.NewHelper(logger).Info("启动定时任务调度器...")
			return scheduler.Start(ctx)
		}),
		// 添加停止钩子来停止 cron 调度器
		kratos.AfterStop(func(ctx context.Context) error {
			log.NewHelper(logger).Info("停止定时任务调度器...")
			return scheduler.Stop(ctx)
		}),
	)

	return app
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
			fmt.Printf("failed to close config: %v\n", err)
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
			fmt.Printf("failed to init tracer: %v\n", err)
		} else {
			defer tracerCleanup()
			fmt.Printf("tracer initialized: %s\n", bc.Trace.Endpoint)
		}
	}

	app, cleanup, err := wireApp(&bc, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
