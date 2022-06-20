package middlewares

import (
	rawHttp "net/http"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/pkg/errors"
)

func NewV1InternalRedirect(srv *http.Server, sur biz.SaveAccountRepo, ident biz.IdentRepo,
	tp biz.ThirdPartyRepo,
	errorHandler ErrHandlerFunc,
	logger *log.Helper) {
	r := srv.Route("/v1")
	r.GET("/internal/redirect", func(ctx http.Context) error {
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
		q := ht.Request().URL.Query()
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
		sessionToken, exp, err := ident.GrantToken(ctx, "aacs", sub.UID)
		if err != nil {
			return errorHandler(ctx, errors.WithMessage(err, "生成 aacs Session 失败"))
		}
		c := &rawHttp.Cookie{
			Name:    "x-aacs-token",
			Value:   sessionToken,
			Expires: exp,
		}
		ht.ReplyHeader().Set("Set-Cookie", c.String())
		logger.Debugf("尝试写入 cookie, %s", c.String())
		appInfo, err := tp.GetInfo(ctx, sub.App)
		if err != nil {
			return errorHandler(ctx, errors.WithMessage(err, "获取应用信息失败"))
		}
		callbackUrl, err := appInfo.BuildCallback(time.Unix(ea, 0), token)
		if err != nil {
			return errorHandler(ctx, errors.WithMessage(err, "第三方应用配置有误"))
		}
		rawHttp.Redirect(ctx.Response(), ctx.Request(), callbackUrl, 302)

		return nil
	})
}
