package middlewares

import (
	"context"
	"encoding/json"
	"io"
	rawHttp "net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang/mock/gomock"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	Path string `json:"path"`
}
type tl struct {
	t *testing.T
}

func (t *tl) Log(level log.Level, keyvals ...interface{}) error {
	switch level {
	case log.LevelInfo, log.LevelDebug, log.LevelWarn:
		t.t.Log(keyvals...)
	case log.LevelError:
		t.t.Error(keyvals...)
	case log.LevelFatal:
		t.t.Fatal(keyvals...)
	default:
		t.t.Log(keyvals...)
	}
	return nil
}

func newTestLogger(t *testing.T) log.Logger {
	return &tl{t: t}
}
func TestNewAuthCallbackServ(t *testing.T) {
	ctl := gomock.NewController(t)
	surRepo := biztest.NewMockSaveAccountRepo(ctl)
	tokenIdentRepo := biztest.NewMockIdentTokenRepo(ctl)
	fn := func(w rawHttp.ResponseWriter, r *rawHttp.Request) {
		_ = json.NewEncoder(w).Encode(testData{Path: r.RequestURI})
	}
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)

	var opts []http.ServerOption

	srv := http.NewServer(opts...)
	srv.HandleFunc("/index", fn)
	srv.HandleFunc("/debug", func(w rawHttp.ResponseWriter, r *rawHttp.Request) {
		_ = json.NewEncoder(w).Encode(testData{Path: r.RequestURI})
	})

	if e, err := srv.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	}

	go func() {
		if err := srv.Start(ctx); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	NewAuthCallbackServ(srv, surRepo, tokenIdentRepo, "/callback", "/debug", func(c http.Context, err error) error {
		return c.JSON(200, testData{Path: "失败了"})
	}, log.NewHelper(newTestLogger(t)))

	e, err := srv.Endpoint()
	require.NoError(t, err)
	client := rawHttp.Client{
		Transport: nil,
		CheckRedirect: func(req *rawHttp.Request, via []*rawHttp.Request) error {
			return rawHttp.ErrUseLastResponse
		},
		Timeout: time.Second,
	}
	t.Run("callback_success", func(t *testing.T) {
		token := "this is token"
		tokenIdentRepo.EXPECT().VerifyToken(gomock.Any(), token).Return(biz.Sub{
			UID:         "",
			DisplayName: "",
			Email:       "",
			PhoneNo:     "",
			Source:      "",
			App:         "",
			Retired:     false,
		}, nil)
		surRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
		tp := biz.ThirdPartyInfo{
			CallbackUrl: e.String() + "/callback",
		}
		reqURL, err := tp.BuildCallback(time.Now().Add(time.Minute), token)
		require.NoError(t, err)
		req, err := rawHttp.NewRequest(rawHttp.MethodGet, reqURL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		location := resp.Header.Get("Location")
		setCookie := resp.Header.Get("set-cookie")
		t.Log("Location: ", location)
		t.Log("Set-Cookie: ", setCookie)
		assert.Equal(t, rawHttp.StatusFound, resp.StatusCode, "http status not 302")
		content, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		t.Log("body ", string(content))
		assert.Contains(t, string(content), "/debug")
	})
	//testClient(t, srv)
	_ = srv.Stop(ctx)
}
func testClient(t *testing.T, srv *http.Server) {

	//for _, test := range tests {
	//	var res testData
	//	err := client.Invoke(context.Background(), test.method, test.path, nil, &res)
	//	if test.path == "/index/notfound" && err != nil {
	//		if e, ok := err.(*errors.Error); ok && e.Code == rawHttp.StatusNotFound {
	//			continue
	//		}
	//	}
	//	if err != nil {
	//		t.Fatalf("invoke  error %v", err)
	//	}
	//	if res.Path != test.path {
	//		t.Errorf("expected %s got %s", test.path, res.Path)
	//	}
	//}
}
