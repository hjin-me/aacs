package middlewares

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	v1 "github.com/lunzi/aacs/api/identification/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type identServ struct {
	v1.UnimplementedIdentificationServer
}

func (i *identServ) Basic(ctx context.Context, request *v1.BasicRequest) (*v1.AuthReply, error) {
	//TODO implement me
	panic("implement me")
}

func (i *identServ) VerifyToken(ctx context.Context, request *v1.TokenRequest) (*v1.TokenInfoReply, error) {
	md := metautils.ExtractIncoming(ctx)
	x := md.Get(authorizationKey)
	ci, ok := FromContext(ctx)
	return &v1.TokenInfoReply{
		Uid:         ci.UID,
		DisplayName: x,
		Email:       "",
		PhoneNo:     ci.AppId,
		Retired:     ok,
	}, nil
}

func TestJWT(t *testing.T) {
	ctl := gomock.NewController(t)
	tpRepo := biztest.NewMockThirdPartyRepo(ctl)
	tpRepo.EXPECT().GetInfo(gomock.Any(), "1111").Return(biz.ThirdPartyInfo{SecretKey: "2222"}, nil)

	nl := newLocalListener()
	s := grpc.NewServer(grpc.UnaryInterceptor(Server(tpRepo)))
	v1.RegisterIdentificationServer(s, &identServ{})
	t.Cleanup(func() {
		s.Stop()
	})
	go func() {
		err := s.Serve(nl)
		if err != nil {
			return
		}
	}()
	conn, err := grpc.Dial(nl.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(),
		grpc.WithUnaryInterceptor(Client("1111", []byte("2222"))),
	)
	require.NoError(t, err)
	ic := v1.NewIdentificationClient(conn)
	token, err := ic.VerifyToken(context.Background(), &v1.TokenRequest{
		Token: "1111",
		App:   "2222",
	})
	require.NoError(t, err)
	assert.Empty(t, token.Uid)
	assert.Equal(t, "1111", token.PhoneNo)
	assert.True(t, token.Retired)
	//t.Log(token.Uid, token.Retired)
	//assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMTExIn0.5hfUwOOhg6JZFQ7ONptSuMUGHzKFlb0PYcuhTPyokTs", token.DisplayName)

}
func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}
