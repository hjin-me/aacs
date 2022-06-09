package dbmodel

import "github.com/uptrace/bun"

type AccountRelation struct {
	bun.BaseModel `bun:"table:account_relations"`
	IdentSource   string `bun:"ident_source,pk,notnull"`
	IdentID       string `bun:"ident_id,pk,notnull"`
	UnityID       string `bun:"unity_id,notnull"`
}
