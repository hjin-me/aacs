package db

import (
	"database/sql"

	"github.com/lunzi/aacs/internal/conf"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bunotel"
)

func NewPG(data *conf.Data) *bun.DB {
	conn := pgdriver.NewConnector(pgdriver.WithDSN(data.GetPg().GetDsn()))
	pgdb := sql.OpenDB(conn)
	// Create a Bun db on top of it.
	db := bun.NewDB(pgdb, pgdialect.New())
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName(conn.Config().Database)))
	return db
}
