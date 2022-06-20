package authmw

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	v1 "github.com/lunzi/aacs/api/identification/v1"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/lunzi/aacs/pkgs/authutils"
)

func Session(rpc v1.IdentificationClient, appId authutils.AppId, logger log.Logger) middleware.Middleware {
	ident := authutils.NewIdent(rpc, appId)
	return middlewares.Session(log.NewHelper(logger), ident)
}

var GetSession = middlewares.GetSession
var GetUID = middlewares.GetUID
var GetToken = middlewares.GetToken

var ClientInfoFromContext = middlewares.FromContext
var ContextWithClientInfo = middlewares.NewContext

type ClientInfo = middlewares.ClientInfo
