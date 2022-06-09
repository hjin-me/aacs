package middlewares

import (
	"context"
	rawHttp "net/http"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lunzi/aacs/internal/biz"
)

const NameUid = biz.NameUid
const NameTk = biz.NameTk
const NameExpiredAt = biz.NameExpiredAt
const NameCookie = biz.NameCookie

// Session is middleware server-side metadata.
func Session(logger *log.Helper, ident biz.IdentTokenRepo) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 注册登陆的回调验证
			if tr, ok := transport.FromServerContext(ctx); ok {
				logger.Debug("请求 ", tr.Operation())
				if tr.Kind() == transport.KindHTTP {
					if ht, ok := tr.(http.Transporter); ok {
						q := ht.Request().URL.Query()
						token := q.Get(NameTk)
						ea, err := strconv.ParseInt(q.Get(NameExpiredAt), 10, 64)
						if err == nil {
							defer func() {
								c := &rawHttp.Cookie{
									Name:    "x-aacs-token",
									Value:   token,
									Expires: time.Unix(ea, 0),
								}
								ht.ReplyHeader().Set("Set-Cookie", c.String())
								logger.Debugf("尝试写入 cookie, %s", c.String())
							}()
						}
					}
				}
			}
			// 下面这一段是身份信息获取
			md, ok := metadata.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			token := handleSession(md, logger)
			if token != "" {
				sub, err := ident.VerifyToken(ctx, token)
				if err != nil {
					logger.Warnf("验证用户token失败, token[%s], %v", token, err)
					return handler(ctx, req)
				}
				ci := ClientInfo{
					AppId: sub.App,
					UID:   sub.UID,
					Token: token,
				}
				ctx = NewContext(ctx, ci)
				logger.Debugf("获取到用户信息, %v", ci)
			}
			return handler(ctx, req)
		}
	}
}
func handleSession(md metadata.Metadata, logger *log.Helper) (token string) {
	ck := md.Get("Cookie")
	defer delete(md, "Cookie")
	if ck != "" {
		header := rawHttp.Header{}
		header.Add("Cookie", ck)
		request := rawHttp.Request{Header: header}
		if c, err := request.Cookie(NameCookie); err == nil {
			token = c.Value
			return token
		} else {
			logger.Debug("cookies 有问题¬")
		}
	}
	a := md.Get("Authorization")
	defer delete(md, "Authorization")
	if len(a) < 8 {
		return
	}
	//Bearer
	return a[7:]
}
