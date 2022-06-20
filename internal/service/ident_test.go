package service

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	v1 "github.com/lunzi/aacs/api/identification/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/data/myotel/myoteltest"
	"github.com/lunzi/aacs/internal/data/pfsession"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIdentificationService(t *testing.T) {
	t.Run("basic_and_verify_token", func(t *testing.T) {
		ctl := gomock.NewController(t)
		identRepo := biztest.NewMockIdentRepo(ctl)
		tpRepo := biztest.NewMockThirdPartyRepo(ctl)
		spanRepo := myoteltest.NewMockSpanExporter(ctl)
		logger := log.NewStdLogger(os.Stdout)
		pfs := pfsession.NewPfSession(context.Background(), defaultDBConf(), spanRepo, log.NewStdLogger(os.Stdout))
		service := NewIdentificationService(identRepo, tpRepo, pfs, logger)

		tpRepo.EXPECT().GetInfo(gomock.Any(), "22").Return(biz.ThirdPartyInfo{
			Name:        "",
			SecretKey:   "",
			CallbackUrl: "",
		}, nil)
		identRepo.EXPECT().Basic(gomock.Any(), "11", "22", "33", "44").Return(biz.Sub{
			UID:         "33",
			DisplayName: "",
			Email:       "",
			PhoneNo:     "",
			Source:      "",
			App:         "22",
			Retired:     false,
		}, nil)
		expired := time.Now().Add(5 * time.Second)
		identRepo.EXPECT().GrantToken(gomock.Any(), "22", "33").Return("123", expired, nil)
		r, err := service.Basic(context.Background(), &v1.BasicRequest{
			Source: "11",
			App:    "22",
			Uid:    "33",
			Pwd:    "44",
		})
		require.NoError(t, err)
		assert.Equal(t, "123", r.GetToken())
		assert.Equal(t, expired.Unix(), r.GetExpiredAt().AsTime().Unix())

		identRepo.EXPECT().VerifyToken(gomock.Any(), r.GetToken()).Return(biz.Sub{
			UID:         "uid123",
			DisplayName: "dn123",
			Email:       "email123",
			PhoneNo:     "phone123",
		}, nil)

		ctx := context.Background()
		ctx = middlewares.NewContext(ctx, middlewares.ClientInfo{
			AppId: "22",
			UID:   "zcvjklz",
			Token: r.GetToken(),
		})
		r2, err := service.VerifyToken(ctx, &v1.TokenRequest{
			Token: r.GetToken(),
			App:   "22",
		})
		require.NoError(t, err)
		assert.Equal(t, "uid123", r2.GetUid())

		ctx = context.Background()
		ctx = middlewares.NewContext(ctx, middlewares.ClientInfo{
			AppId: "not-equal",
			UID:   "zcvjklz",
			Token: r.GetToken(),
		})
		r2, err = service.VerifyToken(ctx, &v1.TokenRequest{
			Token: r.GetToken(),
			App:   "22",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "应用授权不匹配")
	})
	t.Run("who_am_i", func(t *testing.T) {

	})

	t.Run("StandardizeAccount", func(t *testing.T) {
		ctl := gomock.NewController(t)
		identRepo := biztest.NewMockIdentRepo(ctl)
		tpRepo := biztest.NewMockThirdPartyRepo(ctl)
		spanRepo := myoteltest.NewMockSpanExporter(ctl)
		logger := log.NewStdLogger(os.Stdout)
		pfs := pfsession.NewPfSession(context.Background(), defaultDBConf(), spanRepo, log.NewStdLogger(os.Stdout))
		service := NewIdentificationService(identRepo, tpRepo, pfs, logger)
		identRepo.EXPECT().GetUIDByRelation(gomock.Any(), "wecom", "id").Return(biz.Sub{
			UID:         "wecom:id",
			DisplayName: "dn",
			Email:       "em",
			PhoneNo:     "pn",
			Source:      "",
			App:         "",
			Retired:     true,
			Gender:      "",
		}, nil)
		ctx := context.Background()
		ctx = middlewares.NewContext(ctx, middlewares.ClientInfo{
			AppId: "app",
			UID:   "id",
			Token: "mytoken",
		})

		et, err := service.StandardizeAccount(ctx, &v1.StandardizeAccountReq{
			Source: "wecom",
			Id:     "id",
		})
		require.NoError(t, err)
		assert.Equal(t, "wecom:id", et.Uid)
		assert.Empty(t, et.Gender)
		assert.Equal(t, "dn", et.DisplayName)
	})
}
func TestRedirectUrl(t *testing.T) {
	u, err := url.Parse("/internal/redirect")
	require.NoError(t, err)
	values := u.Query()
	values.Set("token", "123#$%&**zxcvas-_")
	u.RawQuery = values.Encode()
	assert.Equal(t, "/internal/redirect?token=123%23%24%25%26%2A%2Azxcvas-_", u.String())
}
func defaultDBConf() *redis.Options {

	viper.SetDefault("REDIS_DSN", "redis://127.0.0.1:6379/0")
	viper.AutomaticEnv()
	dsn := viper.GetString("REDIS_DSN")
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		panic(err)
	}
	return &redis.Options{
		Addr:     opt.Addr,
		Password: opt.Password,
		DB:       opt.DB,
	}
}
