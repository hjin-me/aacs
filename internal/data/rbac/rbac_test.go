package rbac

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/dbtestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tl struct {
	t *testing.T
}

func (t *tl) Log(level log.Level, keyvals ...interface{}) error {
	switch level {
	case log.LevelInfo, log.LevelDebug, log.LevelWarn:
		t.t.Log(keyvals...)
	case log.LevelError:
		t.t.Error(keyvals...)
	case log.LevelFatal:
		t.t.Fatal(keyvals...)
	default:
		t.t.Log(keyvals...)
	}
	return nil
}

func newTestLogger(t *testing.T) log.Logger {
	return &tl{t: t}
}

func TestAuthRepo(t *testing.T) {
	pgDB := dbtestutils.GetRandomPg(t, false)
	a, err := newAuthRepo(pgDB, "aacs", log.NewHelper(newTestLogger(t)))
	require.NoError(t, err)

	t.Run("normal", func(t *testing.T) {
		a.ReplaceLogger(log.NewHelper(newTestLogger(t)))
		ok, err := a.Enforce(
			"aa",
			"bb",
			"cc",
			biz.ActModify,
		)
		require.NoError(t, err)
		assert.False(t, ok)
		err = a.AddUserPolicy(
			"aa",
			[]string{"bb"},
			"cc",
			[]biz.Actions{biz.ActModify},
		)
		require.NoError(t, err)
		ok, err = a.Enforce(
			"aa",
			"bb",
			"cc",
			biz.ActModify,
		)
		require.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("domain_mask", func(t *testing.T) {
		a.ReplaceLogger(log.NewHelper(newTestLogger(t)))
		var ok bool
		var err error
		err = a.AddRolePolicy("domain", []string{"role_huangjin"}, "obj", []biz.Actions{biz.ActRead, biz.ActCreate})
		require.NoError(t, err)
		_, err = a.AddRoleForUser("songsong", "role_huangjin", "*")
		require.NoError(t, err)
		ok, err = a.Enforce("domain", "songsong", "obj", biz.ActCreate)
		require.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("superuser", func(t *testing.T) {
		a.ReplaceLogger(log.NewHelper(newTestLogger(t)))
		var ok bool
		var err error
		err = a.AddRolePolicy("domain", []string{"role_super"}, "*", []biz.Actions{biz.ActRead, biz.ActCreate})
		require.NoError(t, err)
		_, err = a.AddRoleForUser("super_songsong", "role_super", "*")
		require.NoError(t, err)
		t.Log(a.E().GetDomainsForUser("u/super_songsong"))

		ok, err = a.Enforce("domain", "super_songsong", "a", biz.ActCreate)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("sa", func(t *testing.T) {
		a.ReplaceLogger(log.NewHelper(newTestLogger(t)))
		var ok bool
		var err error

		ok, err = a.E().Enforce("r/sa", "aacs", "n:haha", string(biz.ActCreate))
		require.NoError(t, err)
		assert.True(t, ok)
		ok, err = a.Enforce("aacs", "local:root", "haha", biz.ActCreate)
		require.NoError(t, err)
		assert.True(t, ok)

		t.Log(a.GetPermissionsForUser("aacs", "local:root"))

		roles := a.GetRolesForUser("aacs", "local:root")
		assert.Contains(t, roles, "sa")
		assert.Contains(t, roles, "local:root")
		assert.NotContains(t, roles, "r/sa")
		assert.NotContains(t, roles, "u/local:root")

	})
}

func TestNewAuthRepo(t *testing.T) {
	_, err := NewAuthRepo(&conf.Data{Pg: &conf.Data_PG{}}, &conf.Server{RootAppId: "aacs"}, newTestLogger(t))
	require.Error(t, err)
}

func TestCleanPrefix(t *testing.T) {
	cases := []struct {
		expect string
		raw    string
	}{
		{"as", "u/as"},
		{"as", "r/as"},
	}
	for _, c := range cases {
		assert.Equal(t, c.expect, cleanPrefix(c.raw))
	}
}
