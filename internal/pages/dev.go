package pages

import (
	"context"
	rawHttp "net/http"

	"github.com/lunzi/aacs/internal/biz"
	"github.com/pkg/errors"
)

func DevPage(ident biz.IdentRepo, tpRepo biz.ThirdPartyRepo) PageHandler {

	return func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error) {

		tk := r.URL.Query().Get(biz.NameTk)
		if tk == "" {
			return errors.New("未授权的访问")
		}
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
		return Render(ctx, "dev", w, nil)

	}
}
