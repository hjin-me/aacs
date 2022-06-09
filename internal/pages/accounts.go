package pages

import (
	"context"
	rawHttp "net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/server/middlewares"
)

func AccountsLogin(tpRepo biz.ThirdPartyRepo,
	identRepo biz.IdentRepo,
	rootAppId string,
	logger *log.Helper) PageHandler {

	return func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error) {
		app := r.URL.Query().Get("app")
		if app == "" {
			app = rootAppId
		}
		appInfo, err := tpRepo.GetInfo(ctx, app)
		if err != nil {
			return err
		}
		if appInfo.AutoLogin {
			uid, err := middlewares.GetUID(ctx)
			logger.Debug("after auto login ", uid, err)
			if err == nil {
				// grant new token
				tk, expiredAt, err := identRepo.GrantToken(ctx, appInfo.Id, uid)
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
			} else {
				logger.Debug("没有获取到session", err)
			}
		}
		loginStruct := struct {
			Title         string
			App           string
			Url           string
			Copyright     string
			DefaultSource string
		}{
			Title:         appInfo.Name,
			App:           app,
			DefaultSource: "antiy",
		}
		return Render(ctx, "accounts", w, loginStruct)

	}
}
