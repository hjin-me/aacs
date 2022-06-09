package authconn

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestX(t *testing.T) {
	require.Panics(t, func() {
		Client("", nil)
	})
}
