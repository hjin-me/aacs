package dbmodel

import "github.com/uptrace/bun"

type Resource struct {
	bun.BaseModel `bun:"table:openid_idents"`
	Id            int    `bun:"id,pk,autoincrement"`
	Name          string `bun:"name,unique"`
	IsPrimary     bool   `bun:"is_primary"`
	Provider      string `bun:"provider"`
}
