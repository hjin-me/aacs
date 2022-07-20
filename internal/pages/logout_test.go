package pages

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/golang/mock/gomock"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/server/servtest"
)

func TestLogoutGet(t *testing.T) {
	p := &Pages{}
	ctl := gomock.NewController(t)
	tp := biztest.NewMockThirdPartyRepo(ctl)
	tp.EXPECT().GetInfo(gomock.Any(), "aacs").Return(biz.ThirdPartyInfo{
		Id:                "aacs",
		Name:              "AACS",
		SecretKey:         "123123123",
		CallbackUrl:       "http://abc.com/aacs/redirect",
		KeyValidityPeriod: 3600,
		AutoLogin:         true,
		DevMode:           true,
	}, nil)

	srv, expect := servtest.NewHttpTestServer(t)
	r := srv.Route("/")
	r.GET("/logout", p.WrapHandler(LogoutGet(tp)))
	expect.GET("/logout").WithQuery("app", "aacs").WithRedirectPolicy(httpexpect.DontFollowRedirects).Expect().
		Status(http.StatusFound).Body().Contains("http://abc.com")

}
