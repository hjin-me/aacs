package data

import (
	"github.com/google/wire"
	"github.com/lunzi/aacs/internal/data/accounts"
	"github.com/lunzi/aacs/internal/data/db"
	"github.com/lunzi/aacs/internal/data/ident"
	"github.com/lunzi/aacs/internal/data/ident/localp"
	"github.com/lunzi/aacs/internal/data/myotel"
	"github.com/lunzi/aacs/internal/data/pfsession"
	"github.com/lunzi/aacs/internal/data/rbac"
	"github.com/lunzi/aacs/internal/data/thirdparty"
	"github.com/lunzi/aacs/internal/data/wecom"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	rbac.NewAuthRepo,
	thirdparty.NewThirdPartyRepo,
	ident.NewIdentRepo,
	db.NewPG,
	pfsession.NewRedis,
	localp.NewPgProvider,
	ident.NewOpenIdSet,
	accounts.NewAccountsRepo,
	pfsession.NewRedisConf,
	pfsession.NewPfSession,
	myotel.NewTracerClient,
	myotel.NewTracerExporter,
	myotel.NewMetricClient,
	myotel.NewMetricExporter,
	wecom.NewWeCom,
	wecom.NewWeComConf,
)
