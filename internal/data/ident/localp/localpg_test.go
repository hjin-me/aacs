package localp

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	openidproviderv1 "github.com/lunzi/aacs/api/openidprovider/v1"
	"github.com/lunzi/aacs/internal/data/dbtestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPgProvider(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		db := dbtestutils.GetRandomDB(t)
		logger := log.With(log.NewStdLogger(os.Stdout))
		p := NewPgProvider(db, logger)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := p.BasicAuth(ctx, &openidproviderv1.BasicAuthReq{Uid: "songsong", Pwd: "mima"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "用户信息有误")

		err = p.Create(ctx, "songsong", "mima", "mingzi", "youxiang", "phone")
		require.NoError(t, err)

		_, err = p.BasicAuth(ctx, &openidproviderv1.BasicAuthReq{Uid: "songsong", Pwd: "cuowumima"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "用户名密码不匹配")

		_, err = p.BasicAuth(ctx, &openidproviderv1.BasicAuthReq{Uid: "songsong", Pwd: "mima"})
		require.NoError(t, err)
	})
	t.Run("search", func(t *testing.T) {
		db := dbtestutils.GetRandomDB(t)
		logger := log.With(log.NewStdLogger(os.Stdout))
		p := NewPgProvider(db, logger)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var err error

		err = p.Create(ctx, "songsong", "mima", "mingzi", "youxiang", "phone")
		require.NoError(t, err)

		s, err := p.SearchUid(ctx, &openidproviderv1.SearchUidReq{Uid: "songsong"})
		require.NoError(t, err)
		assert.Equal(t, "mingzi", s.Sub.DisplayName)

		_, err = p.SearchUid(ctx, &openidproviderv1.SearchUidReq{Uid: "zhaobudao"})
		require.Error(t, err)
	})
}
