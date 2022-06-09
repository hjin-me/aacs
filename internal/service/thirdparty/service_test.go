package thirdparty

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/mock/gomock"
	v1 "github.com/lunzi/aacs/api/thirdparty/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/stretchr/testify/require"
)

func TestNewThirdPartyService(t *testing.T) {
	ctl := gomock.NewController(t)
	logger := log.NewStdLogger(os.Stdout)
	mockThirdPartyRepo := biztest.NewMockThirdPartyRepo(ctl)
	mockAuthRepo := biztest.NewMockAuthRepo(ctl)
	mockIdentRepo := biztest.NewMockIdentRepo(ctl)
	service := NewThirdPartyService(mockThirdPartyRepo, mockAuthRepo, mockIdentRepo, &conf.Server{
		RootAppId: "aacs",
	}, logger)

	// mock
	mockThirdPartyRepo.EXPECT().Add(gomock.Any(), "appId", "name", "owner", gomock.Any(), true).DoAndReturn(
		func(ctx context.Context, appId, appName, owner, callbackUrl string, autoLogin bool) (biz.ThirdPartyInfo, error) {
			return biz.ThirdPartyInfo{
				Id:                appId,
				Name:              appName,
				SecretKey:         "abcdefg",
				CallbackUrl:       callbackUrl,
				KeyValidityPeriod: 3600,
				AutoLogin:         autoLogin,
			}, nil
		})
	ctx := context.Background()
	ctx = middlewares.NewContext(ctx, middlewares.ClientInfo{
		AppId: "app",
		UID:   "ss",
		Token: "token",
	})
	mockAuthRepo.EXPECT().Enforce(gomock.Any(), "ss", "appId", biz.ActCreate).Return(true, nil)

	add, err := service.Add(ctx, &v1.AddRequest{
		Id:          "appId",
		Name:        "name",
		Owner:       "owner",
		CallbackUrl: "http://127.0.0.1",
		AutoLogin:   true,
	})
	require.NoError(t, err)
	_ = add
	mockAuthRepo.EXPECT().Enforce(gomock.Any(), "ss", "appId", biz.ActModify).Return(true, nil)
	mockIdentRepo.EXPECT().GrantTokenWithPeriod(gomock.Any(), "appId", "ss", gomock.Any()).Return("token", time.Now().Add(1000*time.Second), nil)

	r, err := service.GrantToken(ctx, &v1.GrantTokenReq{
		Id:               "appId",
		PeriodOfValidity: 1000,
	})
	require.NoError(t, err)
	_ = r
}
