package auth

import (
	"context"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/mock/gomock"
	v1 "github.com/lunzi/aacs/api/authorization/v1"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthorizationService(t *testing.T) {
	t.Run("enforce", func(t *testing.T) {
		ctl := gomock.NewController(t)
		authRepo := biztest.NewMockAuthRepo(ctl)
		logger := log.NewStdLogger(os.Stdout)
		service := NewAuthorizationService(logger, authRepo)
		authRepo.EXPECT().Enforce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

		ctx := middlewares.NewContext(context.Background(), middlewares.ClientInfo{
			AppId: "我是APP",
			UID:   "我是UID",
			Token: "我是Token",
		})
		r, err := service.Enforce(ctx, &v1.EnforceReq{
			Obj: "hello/world",
			Sub: "song",
			Act: "del",
		})
		require.NoError(t, err)
		assert.True(t, r.GetResult())
	})
	t.Run("add_permission", func(t *testing.T) {
		ctl := gomock.NewController(t)
		authRepo := biztest.NewMockAuthRepo(ctl)
		logger := log.NewStdLogger(os.Stdout)
		service := NewAuthorizationService(logger, authRepo)

		authRepo.EXPECT().Enforce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		authRepo.EXPECT().AddUserPolicy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		ctx := middlewares.NewContext(context.Background(), middlewares.ClientInfo{
			AppId: "我是APP",
			UID:   "我是UID",
			Token: "我是Token",
		})
		r, err := service.AddPermissionForUser(ctx, &v1.AddPermissionForUserReq{
			Uid: "song",
			Obj: "hello/world",
			Act: "del",
		})
		require.NoError(t, err)
		assert.True(t, r.GetResult())
	})
}
