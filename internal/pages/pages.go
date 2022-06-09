package pages

import (
	"context"
	rawHttp "net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/wecom"
)

type PageServ struct {
	tpRepo    biz.ThirdPartyRepo
	identRepo biz.IdentRepo
	logger    *log.Helper
	p         *Pages
	wc        wecom.WeCom
}

func (s *PageServ) BindServ(srv *http.Server, rootAppId string) {
	r := srv.Route("/")
	r.GET("/", s.p.WrapHandler(PageLogin(s.tpRepo, s.identRepo, rootAppId, s.wc, s.logger)))
	r.GET("/debug", s.p.WrapHandler(HomePage))
	r.GET("/wecom-login", s.p.WrapHandler(PageWecomLogin(s.tpRepo, s.identRepo, s.logger)))
	r.GET("/logout", s.p.WrapHandler(func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error) {
		c := &rawHttp.Cookie{
			Name:    "x-aacs-token",
			Value:   "",
			Expires: time.Now().Add(-10000 * time.Second),
		}
		w.Header().Set("Set-Cookie", c.String())
		u := r.Header.Get("Referer")
		rawHttp.Redirect(w, r, u, rawHttp.StatusFound)
		return nil
	}))
	r.GET("/m/thirdparty", s.p.WrapHandler(ManagerPage))
	r.GET("/m/accounts", s.p.WrapHandler(AccountsLogin(s.tpRepo, s.identRepo, rootAppId, s.logger)))

	r.GET("/my/pwd", s.p.WrapHandler(PwdLogin()))
	r.GET("/my/otp-png", s.p.WrapHandler(OtpLogin()))
	r.GET("/dev/callback", s.p.WrapHandler(DevPage(s.identRepo, s.tpRepo)))
}

func (s *PageServ) ErrPage(c http.Context, err error) error {
	return (s.p.WrapHandler(ErrPageBuilder(err)))(c)
}

func NewPageServ(tpRepo biz.ThirdPartyRepo, identRepo biz.IdentRepo, c *conf.Server, wc wecom.WeCom, logger log.Logger) *PageServ {
	p := &Pages{pageDebug: c.PageDebug}
	return &PageServ{
		identRepo: identRepo,
		tpRepo:    tpRepo,
		logger:    log.NewHelper(logger),
		p:         p,
		wc:        wc,
	}
}
