package ident

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/dbtestutils"
	"github.com/lunzi/aacs/internal/data/ident/localp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cwd string

func TestNewOpenIdSet(t *testing.T) {
	t.Run("search", func(t *testing.T) {
		ctx := context.Background()
		//ctl := gomock.NewController(t)
		db := dbtestutils.GetRandomDB(t)
		logger := log.NewStdLogger(os.Stdout)
		localpProviderIns := localp.NewPgProvider(db, logger)
		opSet := NewOpenIdSet(ctx, db, &conf.Data{}, localpProviderIns)
		s, err := opSet.SearchUid(ctx, "local", "root")
		require.NoError(t, err)
		assert.Equal(t, "local", s.Source)
	})
}

func ldapConf(t *testing.T) *conf.Data {
	path := filepath.Join(cwd, "../../..", "configs/config.yaml")
	c := config.New(config.WithSource(file.NewSource(path)))
	err := c.Load()
	require.NoError(t, err)
	var bc conf.Bootstrap
	err = c.Scan(&bc)
	require.NoError(t, err)
	viper.AutomaticEnv()
	ldc := viper.GetString("LDAP_HOST")
	if ldc != "" {
		bc.Data.Ldap.Host = ldc
	}
	t.Log("ldap host is ", bc.Data.Ldap.Host)
	return bc.Data
}
func init() {
	_, f, _, _ := runtime.Caller(0)
	cwd = filepath.Dir(f)
}
