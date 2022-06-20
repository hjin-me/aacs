package servtest

import (
	httpstd "net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func NewHttpTestServer(t *testing.T, opts ...http.ServerOption) (*http.Server, *httpexpect.Expect) {

	srv := http.NewServer(opts...)
	u, err := srv.Endpoint()
	if err != nil || u == nil || strings.HasSuffix(u.Host, ":0") {
		t.Fatal(err)
	}
	testConfiguration := httpexpect.Config{
		BaseURL: u.String(),
		Client: &httpstd.Client{
			Transport: httpexpect.NewBinder(srv),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
	}

	return srv, httpexpect.WithConfig(testConfiguration)
}
