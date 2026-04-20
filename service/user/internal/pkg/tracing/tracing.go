package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer 初始化 OpenTelemetry 鏈路追蹤
func InitTracer(endpoint string, serviceName string, sampleRate float64) (func(), error) {
	if endpoint == "" {
		// 如果沒有配置 endpoint，返回一個空的 shutdown 函數
		return func() {}, nil
	}

	ctx := context.Background()

	// 創建 OTLP HTTP exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// 設置採樣率
	sampler := sdktrace.AlwaysSample()
	if sampleRate < 1.0 {
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRate))
	}

	// 創建資源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("v1.0.0"),
			attribute.String("deployment.environment", "production"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 創建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 設置全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 返回清理函數
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}, nil
}
