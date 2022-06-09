package biz

//go:generate mockgen -destination=./biztest/auth_mock.go -package=biztest -source=auth.go

// 设计一下权限模型字段
// {sys}:a:{object} 管理员
// {sys}:n:{object} 普通用户
// subject 请求主体
// object 访问的资源
// action 行为

type AuthRepo interface {
	AuthEnforceRepo
	AuthMgrRepo
}
type AuthEnforceRepo interface {
	// Enforce 验证API
	Enforce(sys, sub, obj string, act Actions) (bool, error)
}
type AuthMgrRepo interface {
	// 写API

	AddRoleForUser(uid string, role string, sys string) (bool, error)
	AddUserPolicy(sys string, subs []string, obj string, acts []Actions) error
	AddRolePolicy(sys string, subs []string, obj string, acts []Actions) error
	RemoveUserPolicy(sys, uid, obj string, act Actions) error
	RemoveRolePolicy(sys, role, obj string, act Actions) error
	DeleteRoleForUser(sys, uid, role string) error

	AddAdmin(sys string, uid string) error

	// 读API

	HasRoleForUser(sys string, uid, role string) (bool, error)
	GetRolesForUser(sys, uid string) []string
	GetUsersForRole(sys, role string) []string
	GetPermissionsForUser(sys, uid string) []Perm
}
type Perm struct {
	Obj string
	Act Actions
}

type PolicyReq struct {
	Sys      string
	Object   string
	Subjects []string
	Actions  []Actions
}

type Actions string

const ActCreate Actions = "create"
const ActDelete Actions = "delete"
const ActModify Actions = "modify"
const ActRead Actions = "read"

func ActionStrArr(s []string) []Actions {
	r := make([]Actions, len(s))
	for i, s2 := range s {
		r[i] = Actions(s2)
	}
	return r
}
