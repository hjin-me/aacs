package thirdparty

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/api/apierr"
	v1 "github.com/lunzi/aacs/api/thirdparty/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	v1.UnimplementedThirdPartyServer
	tp        biz.ThirdPartyRepo
	rbac      biz.AuthRepo
	ident     biz.IdentRepo
	logger    *log.Helper
	rootAppId string
}

func NewThirdPartyService(tp biz.ThirdPartyRepo, rbac biz.AuthRepo, ident biz.IdentRepo, c *conf.Server, logger log.Logger) *Service {
	if c.RootAppId == "" {
		panic("conf.server.root_app_id 没有设置")
	}
	return &Service{tp: tp, rbac: rbac, logger: log.NewHelper(logger), rootAppId: c.RootAppId, ident: ident}
}
func (s *Service) Add(ctx context.Context, request *v1.AddRequest) (*v1.AddReply, error) {
	_, err := s.enforce(ctx, request.Id, biz.ActCreate)
	if err != nil {
		return &v1.AddReply{}, err
	}
	r, err := s.tp.Add(ctx, request.GetId(), request.GetName(), request.GetOwner(), request.GetCallbackUrl(), request.GetAutoLogin())
	if err != nil {
		return nil, apierr.ErrorWtf("添加应用失败, %s", err.Error())
	}
	return &v1.AddReply{Info: &v1.Info{
		Id:                r.Id,
		Name:              r.Name,
		CallbackUrl:       r.CallbackUrl,
		KeyValidityPeriod: uint64(r.KeyValidityPeriod),
		AutoLogin:         r.AutoLogin,
		Secret:            r.SecretKey,
		DevMode:           false,
	}}, nil
}

func (s *Service) Inspect(ctx context.Context, request *v1.InfoRequest) (*v1.Info, error) {
	_, err := s.enforce(ctx, request.Id, biz.ActRead)
	if err != nil {
		return &v1.Info{}, err
	}
	info, err := s.tp.GetInfo(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	r := v1.Info{
		Id:                info.Id,
		Name:              info.Name,
		CallbackUrl:       info.CallbackUrl,
		KeyValidityPeriod: uint64(info.KeyValidityPeriod),
		AutoLogin:         info.AutoLogin,
		Secret:            info.SecretKey,
		DevMode:           info.DevMode,
	}
	return &r, nil
}

func (s *Service) All(ctx context.Context, _ *v1.AllRequest) (*v1.AllReply, error) {
	_, err := middlewares.GetUID(ctx)
	if err != nil {
		return nil, apierr.ErrorUnauthorized("身份校验异常, %v", err)
	}
	l, err := s.tp.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	r := v1.AllReply{}
	r.Data = make([]*v1.Info, len(l))
	for i, info := range l {
		r.Data[i] = &v1.Info{
			Id:                info.Id,
			Name:              info.Name,
			CallbackUrl:       info.CallbackUrl,
			KeyValidityPeriod: uint64(info.KeyValidityPeriod),
			AutoLogin:         info.AutoLogin,
			Secret:            "**",
			DevMode:           info.DevMode,
		}
	}
	return &r, nil
}

func (s *Service) BindAdmin(ctx context.Context, request *v1.BindAdminRequest) (*v1.ResultReply, error) {
	_, err := s.enforce(ctx, request.Id, biz.ActModify)
	if err != nil {
		return &v1.ResultReply{}, err
	}
	err = s.rbac.AddAdmin(request.GetId(), request.GetUid())
	if err != nil {
		return nil, apierr.ErrorWtf("绑定管理员失败, %s", err.Error())
	}
	return &v1.ResultReply{Result: true}, nil
}

func (s *Service) GrantToken(ctx context.Context, req *v1.GrantTokenReq) (*v1.GrantTokenReply, error) {
	clientInfo, err := s.enforce(ctx, req.Id, biz.ActModify)
	if err != nil {
		return &v1.GrantTokenReply{}, err
	}
	token, expiredAt, err := s.ident.GrantTokenWithPeriod(ctx, req.Id, clientInfo.UID, time.Second*time.Duration(req.PeriodOfValidity))
	if err != nil {
		return &v1.GrantTokenReply{}, apierr.ErrorWtf("创建token失败, %v", err)
	}
	return &v1.GrantTokenReply{
		Token:     token,
		ExpiredAt: timestamppb.New(expiredAt),
	}, nil
}
func (a *Service) enforce(ctx context.Context, obj string, act biz.Actions) (middlewares.ClientInfo, error) {
	clientInfo, ok := middlewares.FromContext(ctx)
	if !ok {
		return clientInfo, apierr.ErrorUnauthorized("未授权的访问")
	}

	tr := otel.Tracer("thirdparty_rbac")
	_, span := tr.Start(ctx, "enforce")
	ok, err := a.rbac.Enforce(
		a.rootAppId, clientInfo.UID,
		obj, act,
	)
	if err != nil {
		span.SetStatus(codes.Error, "鉴权失败")
		span.RecordError(err)
		span.End()
		return clientInfo, apierr.ErrorWtf("验证用户权限失败 %s", err.Error())
	}
	span.SetStatus(codes.Ok, "成功")
	span.End()
	if !ok {
		return clientInfo, apierr.ErrorUnauthorized("当前用户没有权限")
	}
	return clientInfo, nil
}
