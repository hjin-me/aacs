//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package thirdparty

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/mock/gomock"
	"github.com/google/wire"
	"github.com/lunzi/aacs/internal/biz/biztest"
)

func newThirdPartyService(tp *biztest.MockThirdPartyRepo, rbac *biztest.MockAuthRepo,
	ident *biztest.MockIdentRepo,
	logger log.Logger) (*Service, error) {
	return NewThirdPartyService(tp, rbac, ident, nil, logger), nil
}
func initThirdPartyService(*gomock.Controller, log.Logger) (*Service, func(), error) {
	panic(wire.Build(wire.NewSet(
		biztest.NewMockIdentRepo,
		biztest.NewMockAuthRepo,
		biztest.NewMockOpenIDProvider,
		biztest.NewMockOpenIDSet,
		biztest.NewMockThirdPartyRepo,
	), newThirdPartyService))
}
