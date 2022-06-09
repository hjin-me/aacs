package authutils

import (
	"context"
	"errors"

	v1 "github.com/lunzi/aacs/api/identification/v1"
	"github.com/lunzi/aacs/internal/biz"
)

type AppId string
type Sub = biz.Sub
type Ident interface {
	VerifyToken(ctx context.Context, token string) (Sub, error)
}

type ident struct {
	rpc   v1.IdentificationClient
	appId string
}

func (i *ident) VerifyToken(ctx context.Context, token string) (Sub, error) {
	r, err := i.rpc.VerifyToken(ctx, &v1.TokenRequest{
		Token: token,
		App:   i.appId,
	})
	if err != nil {
		return Sub{}, err
	}
	return Sub{
		UID:         r.GetUid(),
		DisplayName: r.GetDisplayName(),
		Email:       r.GetEmail(),
		PhoneNo:     r.GetPhoneNo(),
		Retired:     r.GetRetired(),
	}, nil
}

// NewIdent
// 定义一个 AppId 类型，方便注入
// ```golang
// func NewAppId() AppId {}
// wire.NewSet(NewAppId)
// ```
func NewIdent(rpc v1.IdentificationClient, appId AppId) Ident {
	if appId == "" {
		panic(errors.New("没有设置应用的AppId"))
	}
	return &ident{
		rpc:   rpc,
		appId: string(appId),
	}
}
