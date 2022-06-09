package thirdparty

import (
	"github.com/uptrace/bun"
)

type Model struct {
	bun.BaseModel `bun:"table:thirdparty"`

	ID                string       `bun:"id,pk" json:"Id,omitempty"`
	AppName           string       `bun:"app_name,notnull" json:"appName,omitempty"`
	Secret            string       `bun:"secret,notnull" json:"secret,omitempty"`
	CallbackUrl       string       `bun:"callback_url,notnull" json:"callbackUrl,omitempty"`
	KeyValidityPeriod uint         `bun:"key_validity_period,notnull" json:"keyValidityPeriod,omitempty"`
	AutoLogin         bool         `bun:"auto_login,notnull,default:true" json:"autoLogin,omitempty"`
	DevMode           bool         `bun:"dev_mode,notnull,default:false" json:"devMode,omitempty"`
	AllowedSource     []string     `bun:"allowed_source,array" json:"allowed_source"`
	DeletedAt         bun.NullTime `bun:"deleted_at" json:"deletedAt"`
}
