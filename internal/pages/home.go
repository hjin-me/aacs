package pages

import (
	"context"
	_ "embed"
	rawHttp "net/http"

	"github.com/lunzi/aacs/internal/server/middlewares"
)

func HomePage(ctx context.Context, _ *rawHttp.Request, w rawHttp.ResponseWriter) error {
	uid, token, _ := middlewares.GetSession(ctx)
	return Render(ctx, "home", w, struct {
		Token string
		Uid   string
	}{
		Token: token,
		Uid:   uid,
	})
}
