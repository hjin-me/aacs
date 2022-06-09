package middlewares

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/lunzi/aacs/internal/biz"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
)

type ThirdPartyRepo interface {
	GetInfo(ctx context.Context, appId string) (biz.ThirdPartyInfo, error)
}

func Server(tp ThirdPartyRepo) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		claims, md, err := validateToken(ctx, tp)
		if err != nil {
			return nil, err
		}
		ctx = NewContext(ctx, ClientInfo{
			AppId: claims.Domain,
			UID:   claims.Subject,
		})
		return handler(md.ToIncoming(ctx), req)
	}
}

// FromContext extract auth info from context
func FromContext(ctx context.Context) (c ClientInfo, ok bool) {
	c, ok = ctx.Value(authKey{}).(ClientInfo)
	return
}

func validateToken(ctx context.Context, tp ThirdPartyRepo) (*biz.Claims, metautils.NiceMD, error) {
	ctx, span := otel.Tracer("auth_middleware").Start(ctx, "verify_token")
	defer span.End()

	md := metautils.ExtractIncoming(ctx)
	auths := strings.SplitN(md.Get(authorizationKey), " ", 2)
	md = md.Del(authorizationKey)
	if len(auths) != 2 || !strings.EqualFold(auths[0], bearerWord) {
		return nil, md, ErrMissingJwtToken
	}
	jwtToken := auths[1]
	var claims *biz.Claims
	tokenInfo, err := jwt.ParseWithClaims(jwtToken, &biz.Claims{}, func(token *jwt.Token) (interface{}, error) {
		var ok bool
		// Don't forget to validate the alg is what you expect:
		if _, ok = token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnSupportSigningMethod
		}
		if claims, ok = token.Claims.(*biz.Claims); !ok {
			return nil, fmt.Errorf("unexpected 验证错误")
		}
		if claims.Domain == "" {
			return nil, fmt.Errorf("token 没有颁发对象")
		}
		appInfo, err := tp.GetInfo(ctx, claims.Domain)
		if err != nil {
			return nil, fmt.Errorf("获取token对应的应用失败, %w", err)
		}
		return []byte(appInfo.SecretKey), nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if ok := errors.As(err, &ve); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				return nil, md, ErrTokenInvalid
			case ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0:
				return nil, md, ErrTokenExpired
			case ve.Errors&(jwt.ValidationErrorUnverifiable|jwt.ValidationErrorSignatureInvalid) != 0:
				return nil, md, errors.Unauthorized("请求验证失败", "Token签名错误")
			}
		}
		return nil, md, errors.Unauthorized("请求验证失败", err.Error())
	} else if !tokenInfo.Valid {
		return nil, md, ErrTokenInvalid
	}
	span.SetStatus(codes.Ok, "")
	return claims, md, nil
}
