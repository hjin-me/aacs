package thirdparty

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/mock/gomock"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/data/dbtestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	ctl := gomock.NewController(t)
	db := dbtestutils.GetRandomDB(t)
	authRepo := biztest.NewMockAuthRepo(ctl)
	authRepo.EXPECT().AddAdmin(gomock.Any(), gomock.Any()).Return(nil)
	tpr := NewThirdPartyRepo(log.DefaultLogger, db, authRepo)
	ctx := context.TODO()
	var err error
	_, err = tpr.Add(ctx, "1111", "1111", "huangjin", "", false)
	require.NoError(t, err)
	var rawTp biz.ThirdPartyInfo
	var noCache time.Duration
	{
		start := time.Now()
		for i := 0; i < 100; i++ {
			rawTp, err = tpr.GetInfo(ctx, "1111")
			require.NoError(t, err)
		}
		noCache = time.Now().Sub(start)
	}

	_, err = tpr.ListAll(ctx)

	require.NoError(t, err)
	var cTp biz.ThirdPartyInfo
	var hasCache time.Duration
	{
		start := time.Now()
		for i := 0; i < 100; i++ {
			cTp, err = tpr.GetInfo(ctx, "1111")
			require.NoError(t, err)
		}
		hasCache = time.Now().Sub(start)
	}
	assert.Equal(t, rawTp, cTp)
	assert.Less(t, hasCache, noCache)
}
