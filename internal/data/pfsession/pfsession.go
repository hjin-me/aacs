package pfsession

import (
	"context"
	"fmt"
	rawHttp "net/http"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func NewPfSession(ctx context.Context, ro *redis.Options, traceExp sdktrace.SpanExporter, logger log.Logger) *PfSession {
	db := NewRedis(ro)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(strings.Split(ro.Addr, ":")[0]),
			semconv.DBConnectionStringKey.String(ro.Addr),
			semconv.DBSystemRedis,
		),
	)
	if err != nil {
		panic(errors.WithMessage(err, "初始化 pf session 失败"))
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(1)),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(traceExp)),
	)
	return &PfSession{sessionName: "AntiySSID", db: db,
		tracer: tracerProvider.Tracer("pf"),
		log:    log.NewHelper(logger, log.WithMessageKey("pfsession"))}
}

type PfSession struct {
	sessionName string
	db          *redis.Client
	log         *log.Helper
	tracer      trace.Tracer
}

func (p *PfSession) serialize(uid string) string {
	return fmt.Sprintf(`manager|a:2:{s:8:"username";s:%d:"%s";s:2:"cn";s:%d:"%s";}`, len(uid), uid, len(uid), uid)
}
func (p *PfSession) SetSession(ctx context.Context, uid string) error {
	ctx, span := p.tracer.Start(ctx, "set_session")
	defer span.End()
	md, ok := metadata.FromServerContext(ctx)
	if !ok {
		p.log.Debugf("非kratos的请求")
		return nil
	}
	ck := md.Get("Cookie")

	var sessionKey string
	if ck != "" {
		header := rawHttp.Header{}
		header.Add("Cookie", ck)
		request := rawHttp.Request{Header: header}
		if c, err := request.Cookie(p.sessionName); err == nil {
			sessionKey = c.Value
		}
	}
	if sessionKey == "" {
		p.log.Debugf("没有获取到 session key")
		return nil
	}

	d := p.serialize(uid)
	err := p.db.Set(ctx, sessionKey, d, time.Second*3600*6).Err()
	if err != nil {
		span.SetStatus(codes.Error, "写入缓存失败")
		return errors.WithMessage(err, "写入Session缓存失败")
	}
	span.SetStatus(codes.Ok, "")
	return nil
}
