package biz

import (
	"context"
	"time"

	openidproviderv1 "github.com/lunzi/aacs/api/openidprovider/v1"
)

type Sub struct {
	UID         string
	DisplayName string
	Email       string
	PhoneNo     string
	Source      string
	App         string
	Retired     bool
	Gender      string
}

//go:generate mockgen -destination=./biztest/ident_mock.go -package=biztest -source=ident.go
type IdentRepo interface {
	IdentTokenRepo
	Basic(ctx context.Context, source string, app string, uid string, pwd string) (Sub, error)
	TokenAuth(ctx context.Context, source string, app string, token string) (Sub, error)
	GrantToken(ctx context.Context, app, uid string) (string, time.Time, error)
	GrantTokenWithPeriod(ctx context.Context, app, uid string, p time.Duration) (string, time.Time, error)
	ParseUID(ctx context.Context, uid string) (ns, id string, err error)
	SaveRelation(ctx context.Context, uId, identId, identSource string) error
	GetUIDByRelation(ctx context.Context, identSource, id string) (Sub, error)
}

type IdentTokenRepo interface {
	VerifyToken(ctx context.Context, token string) (Sub, error)
}

type OpenIDSet interface {
	Register(ctx context.Context, o openidproviderv1.OpenIDProviderClient) error
	Get(ctx context.Context, name string) (openidproviderv1.OpenIDProviderClient, bool, error)
	ParseUID(ctx context.Context, uid string) (ns, id string, err error)
	BasicAuth(ctx context.Context, name, uid, pwd string) (Sub, error)
	TokenAuth(ctx context.Context, name, token string) (Sub, string, error)
	SearchUid(ctx context.Context, name, uid string) (Sub, error)
}

type OpenIDProvider interface {
	Name() string
	Basic(ctx context.Context, uid, pwd string) (Sub, error)
	TokenAuth(ctx context.Context, token string) (Sub, string, error)
	SearchUid(ctx context.Context, uid string) (Sub, error)
}
