package auth

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lunzi/aacs/api/apierr"
	v1 "github.com/lunzi/aacs/api/authorization/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/server/middlewares"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// AuthorizationService 服务
type AuthorizationService struct {
	v1.UnimplementedAuthorizationServer
	log     *log.Helper
	rbac    biz.AuthRepo
	roleObj string
}

func (a *AuthorizationService) Enforce(ctx context.Context, req *v1.EnforceReq) (*v1.ResultReply, error) {
	clientInfo, ok := middlewares.FromContext(ctx)
	if !ok {
		return &v1.ResultReply{Result: false}, apierr.ErrorUnauthorized("未授权的访问")
	}
	ok, err := a.rbac.Enforce(
		clientInfo.AppId,
		req.Sub,
		req.Obj,
		biz.Actions(req.Act),
	)
	if err != nil {
		return &v1.ResultReply{Result: false}, apierr.ErrorWtf("验证用户权限失败 %s", err.Error())
	}
	return &v1.ResultReply{Result: ok}, nil
}

func (a *AuthorizationService) AddPermissionForUser(ctx context.Context, req *v1.AddPermissionForUserReq) (*v1.ResultReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActCreate)
	if err != nil {
		return &v1.ResultReply{Result: false}, err
	}
	err = a.rbac.AddUserPolicy(
		clientInfo.AppId,
		[]string{req.Uid},
		req.Obj,
		biz.ActionStrArr([]string{req.Act}),
	)
	if err != nil {
		return &v1.ResultReply{Result: false}, apierr.ErrorWtf("内部错误, %v", err)
	}
	return &v1.ResultReply{Result: true}, nil
}

func (a *AuthorizationService) AddPermissionForRole(ctx context.Context, req *v1.AddPermissionForRoleReq) (*v1.ResultReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActCreate)
	if err != nil {
		return &v1.ResultReply{Result: false}, err
	}

	err = a.rbac.AddRolePolicy(
		clientInfo.AppId,
		[]string{req.Role},
		req.Obj,
		biz.ActionStrArr([]string{req.Act}),
	)
	if err != nil {
		return &v1.ResultReply{Result: false}, apierr.ErrorWtf("内部错误, %v", err)
	}
	return &v1.ResultReply{Result: true}, nil
}

func (a *AuthorizationService) AddRoleForUser(ctx context.Context, req *v1.AddRoleForUserReq) (*v1.ResultReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActCreate)
	if err != nil {
		return &v1.ResultReply{Result: false}, err
	}

	ok, err := a.rbac.AddRoleForUser(req.Uid, req.Role, clientInfo.AppId)
	if err != nil {
		return &v1.ResultReply{Result: false}, apierr.ErrorWtf("添加角色失败, %v", err)
	}
	return &v1.ResultReply{Result: ok}, nil
}

func (a *AuthorizationService) DeletePermissionForUser(ctx context.Context, req *v1.DeletePermissionForUserReq) (*v1.ResultReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActDelete)
	if err != nil {
		return &v1.ResultReply{Result: false}, err
	}

	err = a.rbac.RemoveUserPolicy(clientInfo.AppId,
		req.Uid,
		req.Obj,
		biz.Actions(req.Act),
	)
	if err != nil {
		return &v1.ResultReply{Result: false}, apierr.ErrorWtf("内部错误, %v", err)
	}
	return &v1.ResultReply{Result: true}, nil

}

func (a *AuthorizationService) DeletePermissionForRole(ctx context.Context, req *v1.DeletePermissionForRoleReq) (*v1.ResultReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActDelete)
	if err != nil {
		return &v1.ResultReply{Result: false}, err
	}

	err = a.rbac.RemoveRolePolicy(clientInfo.AppId,
		req.Role,
		req.Obj,
		biz.Actions(req.Act),
	)
	if err != nil {
		return &v1.ResultReply{Result: false}, apierr.ErrorWtf("内部错误, %v", err)
	}
	return &v1.ResultReply{Result: true}, nil
}

func (a *AuthorizationService) DeleteRoleForUser(ctx context.Context, req *v1.DeleteRoleForUserReq) (*v1.ResultReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActDelete)
	if err != nil {
		return &v1.ResultReply{Result: false}, err
	}

	err = a.rbac.DeleteRoleForUser(clientInfo.AppId,
		req.Uid,
		req.Role,
	)
	if err != nil {
		return &v1.ResultReply{Result: false}, apierr.ErrorWtf("内部错误, %v", err)
	}
	return &v1.ResultReply{Result: true}, nil
}

func (a *AuthorizationService) GetRolesForUser(ctx context.Context, req *v1.GetRolesForUserReq) (*v1.GetRolesForUserReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActRead)
	if err != nil {
		return &v1.GetRolesForUserReply{}, err
	}

	perms := a.rbac.GetRolesForUser(clientInfo.AppId, req.Uid)
	return &v1.GetRolesForUserReply{
		Roles: perms,
	}, nil
}

func (a *AuthorizationService) GetUsersForRole(ctx context.Context, req *v1.GetUsersForRoleReq) (*v1.GetUsersForRoleReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActRead)
	if err != nil {
		return &v1.GetUsersForRoleReply{}, err
	}
	users := a.rbac.GetUsersForRole(clientInfo.AppId, req.Role)
	return &v1.GetUsersForRoleReply{
		Uid: users,
	}, nil
}

func (a *AuthorizationService) HasRoleForUser(ctx context.Context, req *v1.HasRoleForUserReq) (*v1.HasRoleForUserReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActRead)
	if err != nil {
		return &v1.HasRoleForUserReply{}, err
	}
	ok, err := a.rbac.HasRoleForUser(clientInfo.AppId, req.Uid, req.Role)
	if err != nil {
		return &v1.HasRoleForUserReply{}, apierr.ErrorWtf("权限校验发生了异常, %s", err.Error())
	}
	return &v1.HasRoleForUserReply{
		Result: ok,
	}, nil
}

func (a *AuthorizationService) GetPermissionsForUser(ctx context.Context, req *v1.GetPermissionsForUserReq) (*v1.GetPermissionsForUserReply, error) {
	clientInfo, err := a.enforce(ctx, a.roleObj, biz.ActRead)
	if err != nil {
		return &v1.GetPermissionsForUserReply{}, err
	}

	perms := a.rbac.GetPermissionsForUser(clientInfo.AppId, req.Uid)
	result := make([]*v1.GetPermissionsForUserReply_Perm, len(perms))
	for i, p := range perms {
		result[i] = &v1.GetPermissionsForUserReply_Perm{
			Obj: p.Obj,
			Act: string(p.Act),
		}
	}
	return &v1.GetPermissionsForUserReply{Perm: result}, nil
}

func (a *AuthorizationService) enforce(ctx context.Context, obj string, act biz.Actions) (middlewares.ClientInfo, error) {
	clientInfo, ok := middlewares.FromContext(ctx)
	if !ok {
		return clientInfo, apierr.ErrorUnauthorized("未授权的访问")
	}

	tr := otel.Tracer("account_rbac")
	_, span := tr.Start(ctx, "enforce")

	ok, err := a.rbac.Enforce(
		clientInfo.AppId, clientInfo.UID,
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

func NewAuthorizationService(logger log.Logger, rbac biz.AuthRepo) *AuthorizationService {
	return &AuthorizationService{
		log:     log.NewHelper(logger),
		rbac:    rbac,
		roleObj: "sys/policy",
	}
}
