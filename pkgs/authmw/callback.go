package authmw

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	v1 "github.com/lunzi/aacs/api/identification/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/lunzi/aacs/pkgs/authutils"
)

type SaveAccountRepo = biz.SaveAccountRepo
type Account = biz.Account

func NewAuthCallbackServ(rpc v1.IdentificationClient, appId authutils.AppId,
	srv *http.Server, sur SaveAccountRepo, urlPrefix, callbackUrl string,
	errHandler middlewares.ErrHandlerFunc,
	logger log.Logger) {
	ident := authutils.NewIdent(rpc, appId)
	middlewares.NewAuthCallbackServ(srv, sur, ident, urlPrefix, callbackUrl, errHandler, log.NewHelper(logger))
}
