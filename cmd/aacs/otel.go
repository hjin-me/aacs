package main

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/internal/conf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func newOtel(ctx context.Context, c *conf.Server, metricExp export.Exporter, traceExp sdktrace.SpanExporter, logger log.Logger) (func(), error) {
	if metricExp == nil || traceExp == nil {
		return func() {

		}, nil
	}
	log.NewHelper(logger).Infof("正在启动 otel... %s %f", c.Otel.Addr, c.Otel.Fraction)
	defer log.NewHelper(logger).Info("otel 已启动")

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			metricExp,
		),
		controller.WithExporter(metricExp),
		controller.WithCollectPeriod(3*time.Second),
	)
	global.SetMeterProvider(pusher)

	err := pusher.Start(ctx)
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(c.RootAppId),
		),
	)
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(float64(c.Otel.Fraction))),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)

	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := tracerProvider.ForceFlush(ctx); err != nil {
			otel.Handle(err)
		}
		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
		// pushes any last exports to the receiver
		if err := pusher.Stop(cxt); err != nil {
			otel.Handle(err)
		}
	}, nil
}
