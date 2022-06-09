package aacs

import (
	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
	v14 "github.com/lunzi/aacs/api/account/v1"
	v13 "github.com/lunzi/aacs/api/authorization/v1"
	v12 "github.com/lunzi/aacs/api/identification/v1"
	v1 "github.com/lunzi/aacs/api/thirdparty/v1"
	"github.com/lunzi/aacs/internal/assets"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/pages"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/lunzi/aacs/internal/server/p8s"
	"github.com/lunzi/aacs/internal/service"
	"github.com/lunzi/aacs/internal/service/auth"
	"github.com/lunzi/aacs/internal/service/thirdparty"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, identServ *service.IdentificationService,
	tp *thirdparty.Service,
	pages *pages.PageServ,
	authServ *auth.AuthorizationService,
	accountServ *service.AccountService,
	ident biz.IdentRepo,
	accRepo biz.AccountsRepo,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metrics.Server(
				metrics.WithSeconds(prom.NewHistogram(p8s.MetricSeconds)),
				metrics.WithRequests(prom.NewCounter(p8s.MetricRequests)),
			),
			validate.Validator(),
			metadata.Server(
				metadata.WithPropagatedPrefix("")),
			middlewares.Session(log.NewHelper(logger), ident),
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

	srv.Handle("/metrics", promhttp.Handler())
	h := openapiv2.NewHandler()
	// swagger
	srv.HandlePrefix("/q/", h)
	srv.HandlePrefix("/statics/", assets.StaticsServer(c.GetPageDebug()))

	pages.BindServ(srv, c.RootAppId)
	middlewares.NewAuthCallbackServ(srv, accRepo, ident, "/callback", "/debug", pages.ErrPage, log.NewHelper(logger))

	v12.RegisterIdentificationHTTPServer(srv, identServ)
	v1.RegisterThirdPartyHTTPServer(srv, tp)
	v13.RegisterAuthorizationHTTPServer(srv, authServ)
	v14.RegisterAccountHTTPServer(srv, accountServ)
	return srv
}
