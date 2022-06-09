package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/api/account/v1"
	"github.com/lunzi/aacs/api/apierr"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/ident/localp"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

func NewAccountService(logger log.Logger, lp *localp.ProviderIns, ident biz.IdentRepo, rbac biz.AuthRepo,
	accountsRepo biz.AccountsRepo,
	cs *conf.Server) *AccountService {
	if cs.RootAppId == "" {
		panic("应用 rootAppId 没有设置")
	}
	return &AccountService{
		log:       log.NewHelper(logger),
		lp:        lp,
		ident:     ident,
		rbac:      rbac,
		rootAppId: cs.RootAppId,
		accounts:  accountsRepo,
		policy:    "sys/account",
	}
}

type AccountService struct {
	v1.UnimplementedAccountServer
	rootAppId string
	policy    string
	log       *log.Helper
	lp        *localp.ProviderIns
	ident     biz.IdentRepo
	rbac      biz.AuthRepo
	accounts  biz.AccountsRepo
}

func (a *AccountService) Create(ctx context.Context, req *v1.CreateReq) (*emptypb.Empty, error) {
	_, err := a.enforce(ctx, a.policy, biz.ActCreate)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	err = a.lp.Create(ctx, req.Id, req.Pwd, req.DisplayName, req.Email, req.PhoneNo)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorWtf("创建账号失败, %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (a *AccountService) ResetPwd(ctx context.Context, req *v1.ResetPwdReq) (*emptypb.Empty, error) {
	uid, _, err := middlewares.GetSession(ctx)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorUnauthorized("未授权的访问, %v", err)
	}
	if req.NewPwd != req.VerifyPwd {
		return &emptypb.Empty{}, apierr.ErrorResetPwdErr("两次密码不匹配")
	}

	ns, uid, err := a.ident.ParseUID(ctx, uid)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorWtf("重置密码失败了, %v", err)
	}
	// TODO 不要硬编码
	if ns != "local" && ns != "mingdun" {
		return &emptypb.Empty{}, apierr.ErrorWtf("该账号无法在本平台进行重置密码")
	}

	err = a.lp.ResetPwd(ctx, uid, req.OldPwd, req.NewPwd)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorWtf("重置密码失败, %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (a *AccountService) AllAccounts(ctx context.Context, _ *v1.AllAccountsReq) (*v1.AllAccountsReply, error) {
	ci, err := a.enforce(ctx, a.policy, biz.ActRead)
	if err != nil {
		return &v1.AllAccountsReply{}, err
	}
	_ = ci

	bs, err := a.accounts.AllSubject(ctx)
	if err != nil {
		return &v1.AllAccountsReply{}, apierr.ErrorWtf("获取用户列表失败, %v", err)
	}
	acs := make([]*v1.Account, len(bs))
	for i, b := range bs {
		ac := &v1.Account{
			Uid:         b.Id,
			DisplayName: b.DisplayName,
			Email:       b.Email,
			PhoneNo:     b.PhoneNo,
			Retired:     b.Retired,
		}
		ac.RelatedIdents = make([]*v1.Account_Ident, len(b.RelatedIdents))
		for i2, ident := range b.RelatedIdents {
			ac.RelatedIdents[i2] = &v1.Account_Ident{
				Source: ident.Source,
				Id:     ident.Id,
			}
		}

		acs[i] = ac

	}
	return &v1.AllAccountsReply{
		Accounts: acs,
	}, nil
}
func (a *AccountService) SyncWeCom(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	_, err := a.enforce(ctx, a.policy, biz.ActModify)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	err = a.accounts.SyncWecom(ctx)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorWtf("更新企业微信绑定关系失败, %v", err)
	}
	return &emptypb.Empty{}, nil
}
func (a *AccountService) enforce(ctx context.Context, obj string, act biz.Actions) (middlewares.ClientInfo, error) {
	clientInfo, ok := middlewares.FromContext(ctx)
	if !ok {
		return clientInfo, apierr.ErrorUnauthorized("未授权的访问")
	}

	tr := otel.Tracer("account_rbac")
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
func (a *AccountService) ImportAccount(ctx context.Context, req *v1.ImportAccountReq) (*emptypb.Empty, error) {
	ci, err := a.enforce(ctx, a.policy, biz.ActCreate)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorUnauthorized("未授权的访问")
	}
	_ = ci
	acc, err := a.accounts.ImportAccount(ctx, req.Source, req.Uid)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorWtf("导入用户失败, %v", err)
	}
	err = a.ident.SaveRelation(ctx, acc.Id, req.Uid, req.Source)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorWtf("导入用户失败, %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (a *AccountService) SaveRelation(ctx context.Context, req *v1.SaveRelationReq) (*emptypb.Empty, error) {

	ci, err := a.enforce(ctx, a.policy, biz.ActModify)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	_ = ci
	err = a.ident.SaveRelation(ctx, req.Uid, req.IdentId, req.IdentSource)
	if err != nil {
		return &emptypb.Empty{}, apierr.ErrorWtf("更新绑定关系失败, %v", err)
	}
	return &emptypb.Empty{}, nil
}
