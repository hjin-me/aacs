// Package middlewares
// see more in <https://github.com/grpc-ecosystem/go-grpc-middleware/tree/master/auth>
package middlewares

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Option is jwt option.
type Option func(*options)

// Parser is a jwt parser
type options struct {
	signingMethod jwt.SigningMethod
	claims        jwt.Claims
}

type authKey struct{}

const (

	// bearerWord the bearer key word for authorization
	bearerWord string = "Bearer"

	// bearerFormat authorization token format
	bearerFormat string = "Bearer %s"

	// authorizationKey holds the key used to store the JWT Token in the request header.
	authorizationKey string = "Authorization"
)

var (
	ErrMissingJwtToken        = errors.Unauthorized("请求验证失败", "JWT token is missing")
	ErrTokenInvalid           = errors.Unauthorized("请求验证失败", "Token is invalid")
	ErrTokenExpired           = errors.Unauthorized("请求验证失败", "连接身份认证Token过期")
	ErrUnSupportSigningMethod = errors.Unauthorized("请求验证失败", "Wrong signing method")
	ErrSignToken              = errors.Unauthorized("请求验证失败", "Can not sign token.Is the key correct?")
)

// AuthFunc is the pluggable function that performs authentication.
//
// The passed in `Context` will contain the gRPC metadata.MD object (for header-based authentication) and
// the peer.Peer information that can contain transport-based credentials (e.g. `credentials.AuthInfo`).
//
// The returned context will be propagated to handlers, allowing user changes to `Context`. However,
// please make sure that the `Context` returned is a child `Context` of the one passed in.
//
// If error is returned, its `grpc.Code()` will be returned to the user as well as the verbatim message.
// Please make sure you use `codes.Unauthenticated` (lacking auth) and `codes.PermissionDenied`
// (authed, but lacking perms) appropriately.
type AuthFunc func(ctx context.Context) (context.Context, error)

// UnaryClientInterceptor returns a new unary client interceptors that performs per-request auth.
func UnaryClientInterceptor(authFunc AuthFunc) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		ctx, err = authFunc(ctx)
		if err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

type ClientInfo struct {
	AppId string
	UID   string
	Token string
}

// AuthFromMD is a helper function for extracting the :authorization header from the gRPC metadata of the request.
//
// It expects the `:authorization` header to be of a certain scheme (e.g. `basic`, `bearer`), in a
// case-insensitive format (see rfc2617, sec 1.2). If no such authorization is found, or the token
// is of wrong scheme, an error with gRPC status `Unauthenticated` is returned.
func AuthFromMD(ctx context.Context, expectedScheme string) (string, error) {
	val := metautils.ExtractIncoming(ctx).Get(authorizationKey)
	if val == "" {
		return "", status.Errorf(codes.Unauthenticated, "Request unauthenticated with "+expectedScheme)
	}
	splits := strings.SplitN(val, " ", 2)
	if len(splits) < 2 {
		return "", status.Errorf(codes.Unauthenticated, "Bad authorization string")
	}
	if !strings.EqualFold(splits[0], expectedScheme) {
		return "", status.Errorf(codes.Unauthenticated, "Request unauthenticated with "+expectedScheme)
	}
	return splits[1], nil
}
