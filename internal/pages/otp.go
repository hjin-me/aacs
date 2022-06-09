package pages

import (
	"context"
	"image/png"
	rawHttp "net/http"

	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

func OtpLogin() PageHandler {

	return func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error) {
		uid, _, err := middlewares.GetSession(ctx)
		if err != nil {
			return errors.WithMessage(err, "没有登陆")
		}
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "安天",
			AccountName: uid,
		})

		if err != nil {
			return err
		}
		img, err := key.Image(200, 200)
		if err != nil {
			return err
		}
		w.Header().Set("content-type", "image/png")
		err = png.Encode(w, img)
		if err != nil {
			return err
		}

		return nil
	}
}
