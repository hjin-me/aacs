package pages

import (
	"context"
	_ "embed"
	rawHttp "net/http"
)

func ManagerPage(ctx context.Context, _ *rawHttp.Request, w rawHttp.ResponseWriter) error {
	return Render(ctx, "manager", w, nil)
}
