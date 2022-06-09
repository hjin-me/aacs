package authmw

import (
	"context"
	"testing"

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
