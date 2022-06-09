package pages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testDebug struct {
}

func TestContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), &pageDebug{}, "111")
	b := &pageDebug{}
	assert.Equal(t, ctx.Value(b), "111")
	assert.Equal(t, ctx.Value(&pageDebug{}), "111")
	assert.NotEqual(t, ctx.Value(&testDebug{}), ctx.Value(&pageDebug{}))
}
