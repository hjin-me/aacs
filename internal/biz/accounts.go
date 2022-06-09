package biz

import (
	"context"
)

//go:generate mockgen -destination=./biztest/accounts_mock.go -package=biztest -source=accounts.go
type AccountsRepo interface {
	Add(ctx context.Context, a Account, ignoreConflict bool) error
	Save(ctx context.Context, a Account) error
	AllSubject(ctx context.Context) ([]Account, error)
	GetByID(ctx context.Context, id string) (Account, error)
	ImportAccount(ctx context.Context, identSource string, uid string) (Account, error)
	SyncWecom(ctx context.Context) error

	//Revoke2FA()

}
type SaveAccountRepo interface {
	Save(ctx context.Context, a Account) error
}
type Ident struct {
	Source string
	Id     string
}
type Account struct {
	Id            string
	DisplayName   string
	Email         string
	PhoneNo       string
	Retired       bool
	AllowedApps   []string
	RelatedIdents []Ident
}
