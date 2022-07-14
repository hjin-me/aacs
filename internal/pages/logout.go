package pages

import (
	"context"
	rawHttp "net/http"
	"net/url"
	"time"

	"github.com/lunzi/aacs/internal/biz"
	"github.com/pkg/errors"
)

func LogoutGet(tpRepo biz.ThirdPartyRepo) PageHandler {
	return func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error) {
		c := &rawHttp.Cookie{
			Name:    "x-aacs-token",
			Value:   "",
			Expires: time.Now().Add(-10000 * time.Second),
		}
		w.Header().Set("Set-Cookie", c.String())

		u := r.Header.Get("Referer")
		if u != "" {
			rawHttp.Redirect(w, r, u, rawHttp.StatusFound)
			return
		}
		appId := r.URL.Query().Get("app")
		if appId == "" {
			rawHttp.Redirect(w, r, "/", rawHttp.StatusFound)
			return
		}
		tp, err := tpRepo.GetInfo(ctx, appId)
		if err != nil {
			return errors.WithMessage(err, "获取第三方信息失败")
		}
		uu, err := url.Parse(tp.CallbackUrl)
		if err != nil {
			return errors.WithMessagef(err, "解析回调地址失败: %s", tp.CallbackUrl)
		}
		uu.Path = ""
		uu.RawQuery = ""
		rawHttp.Redirect(w, r, uu.String(), rawHttp.StatusFound)

		return nil
	}
}
