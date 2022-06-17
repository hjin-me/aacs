package authmw

import (
	"context"
	"testing"

	"github.com/lunzi/aacs/internal/server/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	ctx := context.Background()
	a, b, _ := GetSession(ctx)
	assert.Empty(t, a)
	assert.Empty(t, b)
	u, _ := GetUID(ctx)
	assert.Empty(t, u)
	tk, _ := GetToken(ctx)
	assert.Empty(t, tk)

	require.Panics(t, func() {
		Session(nil, "", nil)
	})
}

func TestContext(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithClientInfo(ctx, middlewares.ClientInfo{
		AppId: "appId",
		UID:   "uid",
		Token: "token",
	})

	ci, ok := ClientInfoFromContext(ctx)
	require.True(t, ok)
	assert.EqualValues(t, middlewares.ClientInfo{
		AppId: "appId",
		UID:   "uid",
		Token: "token",
	}, ci)
}
