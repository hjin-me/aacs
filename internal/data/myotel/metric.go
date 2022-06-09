package myotel

import (
	"context"

	"github.com/lunzi/aacs/internal/conf"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric/export"
	"google.golang.org/grpc"
)

func NewMetricClient(c *conf.Server) otlpmetric.Client {
	if c == nil || c.Otel == nil || c.Otel.Addr == "" {
		return nil
	}
	metricClient := otlpmetricgrpc.NewClient(
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(c.Otel.Addr),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()))
	return metricClient
}
func NewMetricExporter(ctx context.Context, client otlpmetric.Client) export.Exporter {
	if client == nil {
		return nil
	}
	metricExp, err := otlpmetric.New(ctx, client)
	if err != nil {
		panic(errors.WithMessage(err, "初始化 metric exporter 失败"))
	}
	return metricExp
}
