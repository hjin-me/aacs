package irisv12

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/httptest"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	ctl := gomock.NewController(t)
	ident := biztest.NewMockIdentRepo(ctl)

	ident.EXPECT().VerifyToken(gomock.Any(), "123456").Return(Sub{
		UID:         "uid",
		DisplayName: "dn",
		Email:       "em",
		PhoneNo:     "pn",
		Source:      "s",
		App:         "a",
		Retired:     false,
	}, nil)

	app := iris.New()
	app.Use(Middleware(ident))
	app.Get("/login", func(ct iris.Context) {
		uid, token, err := GetSession(ct.Request().Context())
		require.NoError(t, err)
		assert.NotErrorIs(t, err, ErrNoSession)
		t.Log(uid, token)
		_, _ = ct.WriteString(uid)
	})

	ts := httptest.New(t, app)

	assert.Equal(t, "uid", ts.GET("/login").WithCookie(NameCookie, "123456").Expect().Body().Raw())
	_ = SetSession(context.Background(), ClientInfo{})
}
