package myotel

import (
	"context"

	"github.com/lunzi/aacs/internal/conf"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

func NewTracerClient(c *conf.Server) otlptrace.Client {
	if c == nil || c.Otel == nil || c.Otel.Addr == "" {
		return nil
	}
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(c.Otel.Addr),
		otlptracegrpc.WithDialOption(grpc.WithBlock()))
	return traceClient
}
func NewTracerExporter(ctx context.Context, client otlptrace.Client) trace.SpanExporter {
	if client == nil {
		return nil
	}
	traceExp, err := otlptrace.New(ctx, client)
	if err != nil {
		panic(errors.WithMessage(err, "初始化 tracer exporter 失败"))
	}
	return traceExp
}
