package irisv12

import (
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/pkgs/authutils"
	"github.com/pkg/errors"
)

type ErrHandlerFunc func(c iris.Context, err error)

func AuthCallback(ident authutils.Ident, errorHandler ErrHandlerFunc,
	callbackUrl string,
	sur biz.SaveAccountRepo) iris.Handler {
	return func(c iris.Context) {
		ctx := c.Request().Context()
		q := c.Request().URL.Query()
		token := q.Get(NameTk)
		if token == "" {
			errorHandler(c, errors.New("url参数没有token"))
			return
		}
		ea, err := strconv.ParseInt(q.Get(NameExpiredAt), 10, 64)
		if err != nil {
			errorHandler(c, errors.New("url参数过期日期不正确"))
			return
		}
		sub, err := ident.VerifyToken(ctx, token)
		if err != nil {
			errorHandler(c, errors.New("无效的token"))
			return
		}
		err = sur.Save(ctx, biz.Account{
			Id:          sub.UID,
			DisplayName: sub.DisplayName,
			Email:       sub.Email,
			PhoneNo:     sub.PhoneNo,
			Retired:     sub.Retired,
			AllowedApps: nil,
		})
		if err != nil {
			errorHandler(c, errors.New("用户信息保存失败"))
			return
		}
		cookie := &http.Cookie{
			Name:    "x-aacs-token",
			Value:   token,
			Expires: time.Unix(ea, 0),
		}
		c.SetCookie(cookie)
		c.Redirect(callbackUrl)
	}
}
