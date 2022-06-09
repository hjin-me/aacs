package biztest

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewMockAuthRepo, NewMockIdentRepo)
