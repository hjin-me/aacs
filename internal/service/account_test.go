package service

import (
	"context"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/mock/gomock"
	v1 "github.com/lunzi/aacs/api/account/v1"
	openidproviderv1 "github.com/lunzi/aacs/api/openidprovider/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/dbtestutils"
	"github.com/lunzi/aacs/internal/data/ident/localp"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountService(t *testing.T) {
	t.Run("allAccount", func(t *testing.T) {
		ctl := gomock.NewController(t)
		identRepo := biztest.NewMockIdentRepo(ctl)
		accRepo := biztest.NewMockAccountsRepo(ctl)
		rbacRepo := biztest.NewMockAuthRepo(ctl)
		logger := log.NewStdLogger(os.Stdout)
		db := dbtestutils.GetRandomDB(t)
		lp := localp.NewPgProvider(db, logger)
		service := NewAccountService(logger, lp, identRepo, rbacRepo, accRepo, &conf.Server{RootAppId: "aacs"})
		accRepo.EXPECT().AllSubject(gomock.Any()).Return([]biz.Account{{
			Id:          "Id",
			DisplayName: "DN",
			Email:       "EM",
			PhoneNo:     "PN",
			Retired:     true,
		}}, nil)
		rbacRepo.EXPECT().Enforce("aacs", "我是UID", "sys/account", biz.ActRead).Return(true, nil)

		ctx := middlewares.NewContext(context.Background(), middlewares.ClientInfo{
			AppId: "我是APP",
			UID:   "我是UID",
			Token: "我是Token",
		})
		r, err := service.AllAccounts(ctx, nil)
		require.NoError(t, err)
		require.Len(t, r.Accounts, 1)
		assert.Equal(t, "Id", r.Accounts[0].Uid)

	})
	t.Run("reset", func(t *testing.T) {
		ctl := gomock.NewController(t)
		identRepo := biztest.NewMockIdentRepo(ctl)
		rbacRepo := biztest.NewMockAuthRepo(ctl)
		accRepo := biztest.NewMockAccountsRepo(ctl)
		logger := log.NewStdLogger(os.Stdout)
		db := dbtestutils.GetRandomDB(t)
		lp := localp.NewPgProvider(db, logger)
		service := NewAccountService(logger, lp, identRepo, rbacRepo, accRepo, &conf.Server{RootAppId: "aacs"})
		rbacRepo.EXPECT().Enforce("aacs", "我是UID", "sys/account", biz.ActCreate).Return(true, nil)
		identRepo.EXPECT().ParseUID(gomock.Any(), "我是UID").Return("local", "我是UID", nil)

		ctx := middlewares.NewContext(context.Background(), middlewares.ClientInfo{
			AppId: "我是APP",
			UID:   "我是UID",
			Token: "我是Token",
		})

		var err error
		_, err = service.Create(ctx, &v1.CreateReq{
			Id:          "我是UID",
			DisplayName: "我的名字",
			Email:       "a@a.b",
			PhoneNo:     "182",
			Pwd:         "Old",
		})
		require.NoError(t, err)

		_, err = service.ResetPwd(ctx, &v1.ResetPwdReq{
			OldPwd:    "Old",
			NewPwd:    "New",
			VerifyPwd: "Verify",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "密码不匹配")

		_, err = service.ResetPwd(ctx, &v1.ResetPwdReq{
			OldPwd:    "Old",
			NewPwd:    "New",
			VerifyPwd: "New",
		})
		require.NoError(t, err)

		s, err := lp.BasicAuth(ctx, &openidproviderv1.BasicAuthReq{
			Uid: "我是UID",
			Pwd: "New",
		})
		require.NoError(t, err)
		t.Log(s)
	})

	t.Run("import", func(t *testing.T) {
		ctl := gomock.NewController(t)
		identRepo := biztest.NewMockIdentRepo(ctl)
		rbacRepo := biztest.NewMockAuthRepo(ctl)
		accRepo := biztest.NewMockAccountsRepo(ctl)
		logger := log.NewStdLogger(os.Stdout)
		db := dbtestutils.GetRandomDB(t)
		lp := localp.NewPgProvider(db, logger)
		service := NewAccountService(logger, lp, identRepo, rbacRepo, accRepo, &conf.Server{RootAppId: "aacs"})
		rbacRepo.EXPECT().Enforce("aacs", "我是UID", "sys/account", biz.ActCreate).Return(true, nil)
		rbacRepo.EXPECT().Enforce("aacs", "我没有权限", "sys/account", biz.ActCreate).Return(false, nil)
		accRepo.EXPECT().ImportAccount(gomock.Any(), "local", "我是UID").Return(biz.Account{Id: "我是UID"}, nil)
		identRepo.EXPECT().SaveRelation(gomock.Any(), "我是UID", "我是UID", "local").Return(nil)

		ctx := middlewares.NewContext(context.Background(), middlewares.ClientInfo{
			AppId: "我是APP",
			UID:   "我是UID",
			Token: "我是Token",
		})

		var err error
		_, err = service.ImportAccount(ctx, &v1.ImportAccountReq{
			Source: "local",
			Uid:    "我是UID",
		})
		require.NoError(t, err)

		ctx = middlewares.NewContext(context.Background(), middlewares.ClientInfo{
			AppId: "我是APP",
			UID:   "我没有权限",
			Token: "我是Token",
		})
		_, err = service.ImportAccount(ctx, &v1.ImportAccountReq{
			Source: "local",
			Uid:    "我是UID",
		})
		require.Error(t, err)
	})
	t.Run("SaveRelation", func(t *testing.T) {
		ctl := gomock.NewController(t)
		identRepo := biztest.NewMockIdentRepo(ctl)
		rbacRepo := biztest.NewMockAuthRepo(ctl)
		accRepo := biztest.NewMockAccountsRepo(ctl)
		logger := log.NewStdLogger(os.Stdout)
		db := dbtestutils.GetRandomDB(t)
		lp := localp.NewPgProvider(db, logger)
		service := NewAccountService(logger, lp, identRepo, rbacRepo, accRepo, &conf.Server{RootAppId: "aacs"})
		rbacRepo.EXPECT().Enforce("aacs", "我是UID", "sys/account", biz.ActModify).Return(true, nil)
		rbacRepo.EXPECT().Enforce("aacs", "我没有权限", "sys/account", biz.ActModify).Return(false, nil)
		identRepo.EXPECT().SaveRelation(gomock.Any(), "我是UID", "我是UID", "local").Return(nil)

		ctx := middlewares.NewContext(context.Background(), middlewares.ClientInfo{
			AppId: "我是APP",
			UID:   "我是UID",
			Token: "我是Token",
		})

		var err error
		_, err = service.SaveRelation(ctx, &v1.SaveRelationReq{
			Uid:         "我是UID",
			IdentSource: "local",
			IdentId:     "我是UID",
		})
		require.NoError(t, err)

		ctx = middlewares.NewContext(context.Background(), middlewares.ClientInfo{
			AppId: "我是APP",
			UID:   "我没有权限",
			Token: "我是Token",
		})
		_, err = service.SaveRelation(ctx, &v1.SaveRelationReq{
			Uid:         "我是UID",
			IdentSource: "local",
			IdentId:     "我是UID",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "当前用户没有权限")
	})
}
