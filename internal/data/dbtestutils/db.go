package dbtestutils

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

func GetRandomDB(t *testing.T) *bun.DB {
	// 用于测试，不考虑性能，简单粗暴
	naming := UniqueDBName()
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(DefaultDBConf(""))))

	// Create a Bun db on top of it.
	dbn := bun.NewDB(pgdb, pgdialect.New())
	_, err := dbn.Exec(`create database ` + naming)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := dbn.Exec(`drop database ` + naming)
		require.NoError(t, err)
		_ = dbn.Close()
	})

	db := bun.NewDB(
		sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(DefaultDBConf(naming)))),
		pgdialect.New(),
	)

	t.Cleanup(func() {
		_ = db.Close()
	})

	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	return db
}
func GetRandomPg(t *testing.T, keepDB bool) *pg.DB {
	pgDSN := DefaultDBConf("")
	pgConf, err := pg.ParseURL(pgDSN)
	require.NoError(t, err)
	// 用于测试，不考虑性能，简单粗暴
	naming := UniqueDBName()
	dbn := pg.Connect(pgConf)
	_, err = dbn.Exec(`create database ` + naming)
	require.NoError(t, err)
	t.Log("database name is ", naming)
	if !keepDB {
		t.Cleanup(func() {
			_, err := dbn.Exec(`drop database ` + naming)
			require.NoError(t, err)
			_ = dbn.Close()
		})
	}

	pgConf.Database = naming
	db := pg.Connect(pgConf)
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db

}

func DefaultDBConf(dbname string) string {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgresql://root:123456@127.0.0.1:5432/test?sslmode=disable"
	}
	if dbname != "" {
		dsn = strings.Replace(dsn, "/test?", "/"+dbname+"?", 1)
	}
	return dsn
}
func UniqueDBName() string {
	return "test" + uuid.New().String()[0:8]
}
