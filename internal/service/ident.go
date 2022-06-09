package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/api/apierr"
	v1 "github.com/lunzi/aacs/api/identification/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/data/pfsession"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IdentificationService 服务
type IdentificationService struct {
	v1.UnimplementedIdentificationServer

	ident      biz.IdentRepo
	thirdParty biz.ThirdPartyRepo
	log        *log.Helper
	p          *pfsession.PfSession
}

// NewIdentificationService 初始化
func NewIdentificationService(ident biz.IdentRepo, tp biz.ThirdPartyRepo, p *pfsession.PfSession, logger log.Logger) *IdentificationService {
	return &IdentificationService{ident: ident, thirdParty: tp, p: p, log: log.NewHelper(logger)}
}

func (s *IdentificationService) Basic(ctx context.Context, req *v1.BasicRequest) (*v1.AuthReply, error) {
	// 验证应用权限
	appInfo, err := s.thirdParty.GetInfo(ctx, req.GetApp())
	if err != nil {
		return nil, apierr.ErrorAppInvalid("app[%s] validate failed, %s", req.GetApp(), err.Error())
	}
	// 验证用户名密码数据源（本地/LDAP）
	sub, err := s.ident.Basic(ctx, req.GetSource(), req.GetApp(), req.GetUid(), req.GetPwd())
	if err != nil {
		return nil, apierr.ErrorUserNotFound("用户登陆失败, %s", err.Error())
	}
	token, expired, err := s.ident.GrantToken(ctx, sub.App, sub.UID)
	if err != nil {
		return nil, apierr.ErrorWtf("生成Token失败, %s", err.Error())
	}
	u, err := appInfo.BuildCallback(expired, token)
	if err != nil {
		return nil, apierr.ErrorWtf("第三方应用配置有误, %s", err.Error())
	}
	err = s.p.SetSession(ctx, sub.UID)
	if err != nil {
		return nil, apierr.ErrorWtf("写入安天通用账号失败, %s", err.Error())
	}

	// 返回 access token 和第三方应用的回调页面
	return &v1.AuthReply{
		Token:       token,
		ExpiredAt:   timestamppb.New(expired),
		CallbackUrl: u,
	}, nil
}

func (s *IdentificationService) VerifyToken(ctx context.Context, req *v1.TokenRequest) (*v1.TokenInfoReply, error) {
	ci, ok := middlewares.FromContext(ctx)
	if !ok {
		return &v1.TokenInfoReply{}, apierr.ErrorUnauthorized("未授权的访问")
	}
	if req.App != ci.AppId {
		return &v1.TokenInfoReply{}, apierr.ErrorUnauthorized("应用授权不匹配")
	}
	// 验证应用权限
	token, err := s.ident.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, apierr.ErrorUnauthorized("验证 token 失败, %s", err.Error())
	}
	return &v1.TokenInfoReply{
		Uid:         token.UID,
		DisplayName: token.DisplayName,
		Email:       token.Email,
		PhoneNo:     token.PhoneNo,
		Retired:     token.Retired,
	}, nil
}

func (s *IdentificationService) WhoAmI(ctx context.Context, _ *emptypb.Empty) (*v1.TokenInfoReply, error) {
	_, token, err := middlewares.GetSession(ctx)
	if err != nil {
		return nil, apierr.ErrorUnauthorized("获取当前用户身份失败, %v", err)
	}
	// 验证应用权限
	sub, err := s.ident.VerifyToken(ctx, token)
	if err != nil {
		return nil, apierr.ErrorUnauthorized("验证 token 失败, %s", err.Error())
	}
	return &v1.TokenInfoReply{
		Uid:         sub.UID,
		DisplayName: sub.DisplayName,
		Email:       sub.Email,
		PhoneNo:     sub.PhoneNo,
		Retired:     sub.Retired,
	}, nil
}

func (s *IdentificationService) StandardizeAccount(ctx context.Context, req *v1.StandardizeAccountReq) (*v1.TokenInfoReply, error) {
	_, ok := middlewares.FromContext(ctx)
	if !ok {
		return &v1.TokenInfoReply{}, apierr.ErrorUnauthorized("未授权的访问")
	}
	// 验证应用权限
	sub, err := s.ident.GetUIDByRelation(ctx, req.Source, req.Id)
	if err != nil {
		return nil, apierr.ErrorWtf("查找用户失败, %s", err.Error())
	}
	return &v1.TokenInfoReply{
		Uid:         sub.UID,
		DisplayName: sub.DisplayName,
		Email:       sub.Email,
		PhoneNo:     sub.PhoneNo,
		Retired:     sub.Retired,
		Gender:      sub.Gender,
	}, nil
}
