package pages

import (
	"context"
	rawHttp "net/http"
)

func ErrPageBuilder(err error) PageHandler {
	return func(ctx context.Context, _ *rawHttp.Request, w rawHttp.ResponseWriter) error {
		return Render(ctx, "err", w, struct {
			ErrMsg string
		}{ErrMsg: err.Error()})
	}
}
