package irisv12

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/lunzi/aacs/pkgs/authutils"
)

const NameTk = biz.NameTk
const NameExpiredAt = biz.NameExpiredAt
const NameCookie = biz.NameCookie

var ErrNoSession = middlewares.ErrNoSession

type AppId string
type Sub = biz.Sub
type ClientInfo = middlewares.ClientInfo

func Middleware(ident authutils.Ident) iris.Handler {
	return func(ic iris.Context) {
		ctx := ic.Request().Context()
		q := ic.Request().URL.Query()
		token := q.Get(NameTk)
		ea, err := strconv.ParseInt(q.Get(NameExpiredAt), 10, 64)
		if err == nil {
			defer func() {
				c := &http.Cookie{
					Name:    "x-aacs-token",
					Value:   token,
					Expires: time.Unix(ea, 0),
				}
				ic.SetCookie(c)
			}()
		}

		if token := handleSession(ic); token != "" {
			sub, err := ident.VerifyToken(ctx, token)
			if err != nil {
				ic.Next()
				return
			}
			ci := ClientInfo{
				AppId: sub.App,
				UID:   sub.UID,
				Token: token,
			}
			setSession(ic, ci)
		}

		ic.Next()
	}
}
func handleSession(ic iris.Context) (token string) {
	ck, err := ic.Request().Cookie(NameCookie)
	if err == nil {
		token = ck.Value
		return token
	}
	a := ic.GetHeader("Authorization")
	if len(a) < 8 {
		return
	}
	//Bearer
	return a[7:]
}

func setSession(c iris.Context, ci ClientInfo) {
	ctx := c.Request().Context()
	ctx = middlewares.NewContext(ctx, ci)

	c.ResetRequest(c.Request().WithContext(ctx))
}

func SetSession(ctx context.Context, ci ClientInfo) context.Context {
	return middlewares.NewContext(ctx, ci)
}

func GetSession(ctx context.Context) (uid, token string, err error) {
	return middlewares.GetSession(ctx)
}
