package middlewares

import (
	"encoding/json"
	rawHttp "net/http"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang/mock/gomock"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/server/servtest"
)

func TestNewV1InternalRedirect(t *testing.T) {

	t.Run("callback_success", func(t *testing.T) {

		ctl := gomock.NewController(t)
		surRepo := biztest.NewMockSaveAccountRepo(ctl)
		identRepo := biztest.NewMockIdentRepo(ctl)
		tpRepo := biztest.NewMockThirdPartyRepo(ctl)
		fn := func(w rawHttp.ResponseWriter, r *rawHttp.Request) {
			_ = json.NewEncoder(w).Encode(testData{Path: r.RequestURI})
		}

		srv, expect := servtest.NewHttpTestServer(t)
		srv.HandleFunc("/index", fn)
		srv.HandleFunc("/debug", func(w rawHttp.ResponseWriter, r *rawHttp.Request) {
			_ = json.NewEncoder(w).Encode(testData{Path: r.RequestURI})
		})
		NewV1InternalRedirect(srv, surRepo, identRepo, tpRepo, func(c http.Context, err error) error {
			return c.JSON(200, testData{Path: "失败了"})
		})

		token := "this is token"
		identRepo.EXPECT().VerifyToken(gomock.Any(), token).Return(biz.Sub{
			UID:         "thisisuid",
			DisplayName: "",
			Email:       "",
			PhoneNo:     "",
			Source:      "",
			App:         "bbcs",
			Retired:     false,
		}, nil)
		identRepo.EXPECT().GrantToken(gomock.Any(), "aacs", "thisisuid").Return(
			"newtoken", time.Now().Add(10*time.Minute), nil)
		surRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
		tpRepo.EXPECT().GetInfo(gomock.Any(), "bbcs").Return(biz.ThirdPartyInfo{
			Id:                "bbcs",
			Name:              "",
			SecretKey:         "123123123",
			CallbackUrl:       "/debug",
			KeyValidityPeriod: 3600,
			AutoLogin:         false,
			DevMode:           false,
		}, nil)

		result := expect.GET("/v1/internal/redirect").
			WithRedirectPolicy(httpexpect.DontFollowRedirects).
			WithQuery(NameTk, token).
			WithQuery(NameExpiredAt, time.Now().Add(time.Hour).Unix()).
			Expect()
		result.Status(rawHttp.StatusFound)
		result.Body().Contains("/debug").Contains("this+is+token")
		result.Headers().ContainsKey("Location").ContainsKey("Set-Cookie")
		result.Header("Set-Cookie").Contains("token=newtoken")

	})
}
