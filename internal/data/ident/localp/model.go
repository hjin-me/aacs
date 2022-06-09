package localp

import "github.com/uptrace/bun"

type LocalUser struct {
	bun.BaseModel `bun:"table:local_user"`
	ID            string       `bun:"id,pk"`
	Pwd           string       `bun:"pwd"`
	Salt          string       `bun:"salt"`
	DisplayName   string       `bun:"display_name"`
	Email         string       `bun:"email"`
	PhoneNo       string       `bun:"phone_no"`
	Retired       bool         `bun:"retired"`
	DeletedAt     bun.NullTime `bun:"deleted_at"`
}
