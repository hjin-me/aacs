package dbmodel

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel    `bun:"table:accounts"`
	ID               string             `bun:"id,pk"`
	DisplayName      string             `bun:"display_name"`
	Email            string             `bun:"email"`
	PhoneNo          string             `bun:"phone_no"`
	Retired          bool               `bun:"retired"`
	AllowedApps      []string           `bun:"allowed_apps,array"`
	AccountRelations []*AccountRelation `bun:"rel:has-many,join:id=unity_id"`
	Enable2FA        bool               `bun:"enable_2fa,notnull,default:false"`
	Secret2FA        string             `bun:"secret_2fa,notnull,default:false"`
	CreatedAt        time.Time          `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt        time.Time          `bun:"created_at,nullzero,notnull,default:now()"`
	DeletedAt        bun.NullTime       `bun:"deleted_at"`
}

func (u *Account) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		u.CreatedAt = time.Now()
	case *bun.UpdateQuery:
		u.UpdatedAt = time.Now()
	}
	return nil
}
