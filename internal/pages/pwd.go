package pages

import (
	"context"
	rawHttp "net/http"

	"github.com/lunzi/aacs/internal/server/middlewares"
)

func PwdLogin() PageHandler {

	return func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error) {
		_, _, err = middlewares.GetSession(ctx)
		if err != nil {
			return Render(ctx, "err", w, struct {
				ErrMsg string
			}{ErrMsg: err.Error()})
		}
		return Render(ctx, "pwd", w, nil)

	}
}
