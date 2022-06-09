package myoteltest

//go:generate mockgen -destination=./span_exporter_mock.go -package=myoteltest -source=../../../../vendor/go.opentelemetry.io/otel/sdk/trace/span_exporter.go
//go:generate mockgen -destination=./metric_mock.go -package=myoteltest -source=../../../../vendor/go.opentelemetry.io/otel/sdk/metric/export/metric.go
