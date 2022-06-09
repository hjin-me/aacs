package authmw

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAuthCallbackServ(t *testing.T) {
	require.Panics(t, func() {
		NewAuthCallbackServ(nil, "", nil, nil, "", "", nil, nil)
	})
}
