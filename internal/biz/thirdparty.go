package biz

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const NameUid = "x-aacs-uid"
const NameTk = "x-aacs-token"
const NameExpiredAt = "x-aacs-expired-at"
const NameCookie = NameTk

type ThirdPartyInfo struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	SecretKey         string `json:"secretKey"`
	CallbackUrl       string `json:"callbackUrl"`
	KeyValidityPeriod uint   `json:"keyValidityPeriod"`
	AutoLogin         bool   `json:"autoLogin"`
	DevMode           bool   `json:"devMode"`
}

func (t ThirdPartyInfo) BuildCallback(expiredAt time.Time, token string) (string, error) {
	u, err := url.Parse(t.CallbackUrl)
	if err != nil {
		return "", errors.WithMessage(err, "第三方应用回调地址配置有误")
	}
	values := u.Query()
	values.Set(NameTk, token)
	values.Set(NameExpiredAt, strconv.FormatInt(expiredAt.Unix(), 10))
	u.RawQuery = values.Encode()
	return u.String(), nil
}

//go:generate mockgen -destination=./biztest/thirdparty_mock.go -package=biztest -source=thirdparty.go
type ThirdPartyRepo interface {
	VerifyThirdParty(ctx context.Context, appId string) (bool, error)
	GetInfo(ctx context.Context, appId string) (ThirdPartyInfo, error)
	ListAll(ctx context.Context) ([]ThirdPartyInfo, error)
	Add(ctx context.Context, appId, appName string, owner string, callbackUrl string, autoLogin bool) (ThirdPartyInfo, error)
}
