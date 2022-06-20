package servtest

import (
	"net/http"
	"testing"
)

func TestNewHttpServer(t *testing.T) {
	_, expect := NewHttpTestServer(t)
	expect.POST("/some/path").Expect().Status(http.StatusNotFound)
}
