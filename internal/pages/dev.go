package pages

import (
	"context"
	rawHttp "net/http"
	"net/url"

	"github.com/lunzi/aacs/internal/biz"
	"github.com/pkg/errors"
)

func DevPage(ident biz.IdentRepo, tpRepo biz.ThirdPartyRepo) PageHandler {

	return func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error) {

		tk := r.URL.Query().Get(biz.NameTk)
		if tk == "" {
			return errors.New("未授权的访问")
		}
		expAt := r.URL.Query().Get(biz.NameExpiredAt)
		sub, err := ident.VerifyToken(ctx, tk)
		if err != nil {
			return errors.WithMessage(err, "未授权的访问")
		}
		appInfo, err := tpRepo.GetInfo(ctx, sub.App)
		if err != nil {
			return errors.WithMessage(err, "未授权的访问")
		}
		if !appInfo.DevMode {
			return errors.New("非开发模式")
		}
		u, err := url.Parse(appInfo.CallbackUrl)
		if err != nil {
			return errors.WithMessage(err, "应用回调地址不合法")
		}
		values := u.Query()
		values.Set(biz.NameTk, tk)
		values.Set(biz.NameExpiredAt, expAt)
		u.RawQuery = values.Encode()

		return Render(ctx, "dev", w, struct {
			CallbackUrl string `json:"callbackUrl"`
		}{CallbackUrl: u.String()})

	}
}
