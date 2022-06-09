package middlewares

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/lunzi/aacs/internal/biz"
	"google.golang.org/grpc"
)

// Client is a client jwt middleware.
// 接入说明
// 访问 aacs GRPC 接口时需要做 token 验证，可以访问 查看你有管理权限的应用 id 和 密钥
// 代码示例
// ```
//  import (
//	    v1 "github.com/lunzi/aacs/api/identification/v1"
//      "github.com/lunzi/aacs/pkgs/connauth"
//  )
// 	conn, _ := grpc.Dial("认证服务后端地址",
//		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(),
//		grpc.WithUnaryInterceptor(connauth.Client("应用ID", []byte("应用密钥"))),
//	)
//	return v1.NewIdentificationClient(conn)
// ```
func Client(appId string, key []byte) grpc.UnaryClientInterceptor {
	if len(key) == 0 {
		panic(fmt.Errorf("grpc客户端没有设置密钥"))
	}
	return UnaryClientInterceptor(func(ctx context.Context) (context.Context, error) {
		o := &options{
			signingMethod: jwt.SigningMethodHS256,
			claims: biz.Claims{
				Issuer:    appId,
				Subject:   "",
				Audience:  nil,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
				Version:   "v1",
				Domain:    appId,
			},
		}
		token := jwt.NewWithClaims(o.signingMethod, o.claims)
		tokenStr, err := token.SignedString(key)
		if err != nil {
			return nil, ErrSignToken
		}
		md := metautils.ExtractOutgoing(ctx)
		md = md.Set(authorizationKey, fmt.Sprintf(bearerFormat, tokenStr))
		return md.ToOutgoing(ctx), nil
	})
}

// NewContext put auth info into context
func NewContext(ctx context.Context, c ClientInfo) context.Context {
	return context.WithValue(ctx, authKey{}, c)
}
