package wecom

import (
	"net/url"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/pkg/errors"
)

type Employee struct {
	Name   string `json:"name"`
	UserId string `json:"userid"`
	Mobile string `json:"mobile"`
	Gender string `json:"gender"`
	Email  string `json:"email"`
}

//go:generate mockgen -destination=./wecom_mock.go -package=wecom -source=wecom.go
type WeCom interface {
	LoginUrl(redirectUrl string) (string, error)
}

type Conf struct {
	CorpId string
}

func NewWeCom(c Conf, logger log.Logger) WeCom {
	return &wecom{conf: c, logger: log.NewHelper(logger)}
}

func NewWeComConf(c *conf.Data) Conf {
	return Conf{
		CorpId: c.Wecom.CorpId,
	}
}

type wecom struct {
	conf   Conf
	logger *log.Helper
}

func (q *wecom) LoginUrl(redirectUrl string) (string, error) {
	u, err := url.ParseRequestURI("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=#wechat_redirect")
	if err != nil {
		return "", errors.WithMessage(err, "企业微信登陆地址URL格式错误")
	}

	query := u.Query()
	query.Set("appid", q.conf.CorpId)
	query.Set("redirect_uri", redirectUrl)
	u.RawQuery = query.Encode()
	return u.String(), nil
}
