package aacs

import (
	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	v14 "github.com/lunzi/aacs/api/account/v1"
	v1 "github.com/lunzi/aacs/api/authorization/v1"
	v12 "github.com/lunzi/aacs/api/identification/v1"
	v13 "github.com/lunzi/aacs/api/thirdparty/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/lunzi/aacs/internal/server/p8s"
	"github.com/lunzi/aacs/internal/service"
	"github.com/lunzi/aacs/internal/service/auth"
	"github.com/lunzi/aacs/internal/service/thirdparty"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server,
	ident *service.IdentificationService,
	auth *auth.AuthorizationService,
	tp *thirdparty.Service,
	account *service.AccountService,
	identRepo biz.IdentRepo,
	tpRepo biz.ThirdPartyRepo,
	logger log.Logger,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metrics.Server(
				metrics.WithSeconds(prom.NewHistogram(p8s.MetricSeconds)),
				metrics.WithRequests(prom.NewCounter(p8s.MetricRequests)),
			),
			metadata.Server(
				metadata.WithPropagatedPrefix("")),
			validate.Validator(),
			middlewares.Session(log.NewHelper(logger), identRepo),
		),
		grpc.UnaryInterceptor(middlewares.Server(tpRepo)),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v12.RegisterIdentificationServer(srv, ident)
	v1.RegisterAuthorizationServer(srv, auth)
	v13.RegisterThirdPartyServer(srv, tp)
	v14.RegisterAccountServer(srv, account)
	return srv
}
