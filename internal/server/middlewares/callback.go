package middlewares

import (
	"errors"
	rawHttp "net/http"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lunzi/aacs/internal/biz"
)

type ErrHandlerFunc func(ctx http.Context, err error) error

func NewAuthCallbackServ(srv *http.Server, sur biz.SaveAccountRepo, ident biz.IdentTokenRepo, urlPrefix, callbackUrl string,
	errorHandler ErrHandlerFunc,
	logger *log.Helper) {
	r := srv.Route(urlPrefix)
	r.GET("", func(ctx http.Context) error {
		// 注册登陆的回调验证
		tr, ok := transport.FromServerContext(ctx)
		if !ok {
			return nil
		}
		if tr.Kind() != transport.KindHTTP {
			return nil
		}
		ht, ok := tr.(http.Transporter)
		if !ok {
			return nil
		}
		logger.Debugf("获取 http, %s", ht.Request().URL.RawQuery)
		q := ht.Request().URL.Query()
		logger.Debugf("query, %v, %s", q, q.Get(NameTk))
		token := q.Get(NameTk)
		if token == "" {
			return errorHandler(ctx, errors.New("url参数没有token"))
		}
		ea, err := strconv.ParseInt(q.Get(NameExpiredAt), 10, 64)
		if err != nil {
			logger.Debugf("没有过期时间, %v", err)
			return errorHandler(ctx, errors.New("url参数过期日期不正确"))
		}
		sub, err := ident.VerifyToken(ctx.Request().Context(), token)
		if err != nil {
			logger.Debugf("无效的token, %v", err)
			return errorHandler(ctx, errors.New("无效的token"))
		}
		err = sur.Save(ctx.Request().Context(), biz.Account{
			Id:          sub.UID,
			DisplayName: sub.DisplayName,
			Email:       sub.Email,
			PhoneNo:     sub.PhoneNo,
			Retired:     sub.Retired,
			AllowedApps: nil,
		})
		if err != nil {
			logger.Debugf("保存用户失败, %v", err)
			return errorHandler(ctx, errors.New("用户信息保存失败"))
		}
		c := &rawHttp.Cookie{
			Name:    "x-aacs-token",
			Value:   token,
			Expires: time.Unix(ea, 0),
		}
		ht.ReplyHeader().Set("Set-Cookie", c.String())
		logger.Debugf("尝试写入 cookie, %s", c.String())
		rawHttp.Redirect(ctx.Response(), ctx.Request(), callbackUrl, 302)

		return nil
	})
}
