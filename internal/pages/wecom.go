package pages

import (
	"context"
	_ "embed"
	"errors"
	rawHttp "net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/internal/biz"
)

func PageWecomLogin(tpRepo biz.ThirdPartyRepo,
	identRepo biz.IdentRepo,
	logger *log.Helper) PageHandler {

	return func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error) {
		app := r.URL.Query().Get("app")
		if app == "" {
			return errors.New("缺少参数 app")
		}
		code := r.URL.Query().Get("code")
		if app == "" {
			return errors.New("缺少参数 code")
		}
		appInfo, err := tpRepo.GetInfo(ctx, app)
		if err != nil {
			return err
		}
		sub, err := identRepo.TokenAuth(ctx, "wecom", app, code)
		if err != nil {
			return err
		}
		// grant new token
		tk, expiredAt, err := identRepo.GrantToken(ctx, appInfo.Id, sub.UID)
		if err != nil {
			return err
		}
		cbUrl, err := appInfo.BuildCallback(expiredAt, tk)
		if err != nil {
			return err
		}
		logger.Debug("callback url ", cbUrl)
		rawHttp.Redirect(w, r, cbUrl, 302)

		return nil

	}
}
