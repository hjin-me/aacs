package service

import (
	"github.com/google/wire"
	"github.com/lunzi/aacs/internal/service/auth"
	"github.com/lunzi/aacs/internal/service/thirdparty"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewIdentificationService, auth.NewAuthorizationService,
	thirdparty.NewThirdPartyService, NewAccountService,
)
