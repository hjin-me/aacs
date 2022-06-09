//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data"
	"github.com/lunzi/aacs/internal/pages"
	"github.com/lunzi/aacs/internal/server/aacs"
	"github.com/lunzi/aacs/internal/service"
)

// initApp init kratos application.
func initApp(context.Context, *conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(aacs.ProviderSet, data.ProviderSet, service.ProviderSet, pages.ProviderSet, newApp))
}

func initProvider(context.Context, *conf.Server, log.Logger) (func(), error) {
	panic(wire.Build(data.ProviderSet, newOtel))
}
