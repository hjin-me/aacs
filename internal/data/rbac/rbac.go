package rbac

import (
	_ "embed"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	defaultrolemanager "github.com/casbin/casbin/v2/rbac/default-role-manager"
	"github.com/casbin/casbin/v2/util"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-pg/pg/v10"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	pgadapter "github.com/lunzi/aacs/internal/data/casbin-pg-adapter"
	"github.com/pkg/errors"
)

const UnityPrefix = "u/"
const RolePrefix = "r/"

type authRepo struct {
	e         *casbin.Enforcer
	logger    *log.Helper
	rootAppId string
}

func (a *authRepo) ReplaceLogger(logger *log.Helper) {
	a.logger = logger
}

func (a *authRepo) Enforce(sys, sub, obj string, act biz.Actions) (bool, error) {
	args := []interface{}{
		fmt.Sprintf("%s%s", UnityPrefix, sub),
		sys,
		obj,
		string(act),
	}
	a.logger.Debug("verify\t", args)
	ok, err := a.e.Enforce(args...)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (a *authRepo) VerifyAdmin(sys, sub, obj string, act biz.Actions) (bool, error) {
	ok, err := a.e.Enforce(
		prefix(UnityPrefix, sub),
		sys,
		fmt.Sprintf("a:%s", obj),
		string(act),
	)
	if err != nil {
		return false, errors.WithMessage(err, "验证管理权限失败")
	}
	return ok, nil
}

func (a *authRepo) AddUserPolicy(sys string, subs []string, obj string, acts []biz.Actions) error {
	var rules [][]string
	for _, sub := range subs {
		for _, action := range acts {
			rules = append(rules, []string{
				prefix(UnityPrefix, sub),
				sys,
				obj,
				string(action)},
			)
		}
	}
	ok, err := a.e.AddPolicies(rules)
	if err != nil {
		return errors.WithMessage(err, "添加权限失败")
	}
	if !ok {
		return errors.New("添加权限失败")
	}
	err = a.e.SavePolicy()
	if err != nil {
		return errors.WithMessage(err, "保存权限失败")
	}
	return nil
}

func (a *authRepo) AddAdmin(sys string, uid string) error {
	sub := prefix(UnityPrefix, uid)
	role := prefix(RolePrefix, "sa")
	_, err := a.e.AddRoleForUserInDomain(
		sub,
		role,
		sys,
	)
	if err != nil {
		return errors.WithMessage(err, "添加用户角色失败")
	}
	var rules [][]string
	rules = append(rules, []string{role, sys, "a:*", string(biz.ActCreate)})
	rules = append(rules, []string{role, sys, "a:*", string(biz.ActModify)})
	rules = append(rules, []string{role, sys, "a:*", string(biz.ActDelete)})
	rules = append(rules, []string{role, sys, "a:*", string(biz.ActRead)})

	rules = append(rules, []string{
		sub, a.rootAppId, sys, string(biz.ActCreate)})
	rules = append(rules, []string{
		sub, a.rootAppId, sys, string(biz.ActModify)})
	rules = append(rules, []string{
		sub, a.rootAppId, sys, string(biz.ActDelete)})
	rules = append(rules, []string{
		sub, a.rootAppId, sys, string(biz.ActRead)})
	_, err = a.e.AddPolicies(rules)
	if err != nil {
		return errors.WithMessage(err, "添加权限失败")
	}
	err = a.e.SavePolicy()
	if err != nil {
		return errors.WithMessage(err, "保存权限失败")
	}
	return nil
}

func (a *authRepo) RemoveUserPolicy(sys, uid, obj string, act biz.Actions) error {
	_, err := a.e.RemovePolicy(prefix(UnityPrefix, uid), sys, obj, string(act))
	if err != nil {
		return errors.WithMessage(err, "删除用户权限失败")
	}
	err = a.e.SavePolicy()
	if err != nil {
		return errors.WithMessage(err, "保存权限失败")
	}
	return nil
}

func (a *authRepo) RemoveRolePolicy(sys, role, obj string, act biz.Actions) error {
	_, err := a.e.RemovePolicy(prefix(RolePrefix, role), sys, obj, string(act))
	if err != nil {
		return errors.WithMessage(err, "删除角色权限失败")
	}
	err = a.e.SavePolicy()
	if err != nil {
		return errors.WithMessage(err, "保存权限失败")
	}
	return nil
}

func (a *authRepo) AddRoleForUser(uid string, role string, sys string) (bool, error) {
	//a.e.GetAllRoles()
	ok, err := a.e.AddRoleForUserInDomain(
		prefix(UnityPrefix, uid),
		prefix(RolePrefix, role),
		sys,
	)
	if err != nil {
		return false, errors.WithMessage(err, "添加权限失败")
	}
	err = a.e.SavePolicy()
	if err != nil {
		return ok, errors.WithMessage(err, "保存权限失败")
	}
	return ok, nil
}

func (a *authRepo) AddRolePolicy(sys string, subs []string, obj string, acts []biz.Actions) error {
	var rules [][]string
	for _, sub := range subs {
		for _, action := range acts {
			rules = append(rules, []string{
				prefix(RolePrefix, sub),
				sys,
				obj,
				string(action)})
		}
	}
	ok, err := a.e.AddPolicies(rules)
	if err != nil {
		return errors.WithMessage(err, "添加权限失败")
	}
	if !ok {
		return errors.New("添加权限失败")
	}
	err = a.e.SavePolicy()
	if err != nil {
		return errors.WithMessage(err, "保存权限失败")
	}
	return nil
}

func (a *authRepo) GetPermissionsForUser(sys, uid string) []biz.Perm {
	perms := a.e.GetPermissionsForUserInDomain(prefix(UnityPrefix, uid), sys)
	a.logger.Debug(perms)
	r := make([]biz.Perm, len(perms))
	for i, perm := range perms {
		if len(perm) > 2 {
			r[i] = biz.Perm{
				Obj: perm[2],
			}
		}
		if len(perm) > 3 {
			r[i].Act = biz.Actions(perm[3])
		}
	}
	return r
}

func (a *authRepo) GetRolesForUser(sys, uid string) []string {
	perms := a.e.GetRolesForUserInDomain(prefix(UnityPrefix, uid), sys)
	for i, perm := range perms {
		perms[i] = cleanPrefix(perm)
	}
	return perms
}

func (a *authRepo) DeleteRoleForUser(sys, uid, role string) error {
	ok, err := a.e.DeleteRoleForUserInDomain(prefix(UnityPrefix, uid), prefix(RolePrefix, role), sys)
	if err != nil {
		return errors.WithMessage(err, "删除权限失败")
	}
	if !ok {
		a.logger.Infof("权限删除操作没有影响任何信息 %s %s %s %s", "DeleteRoleForUserInDomain", uid, role, sys)
	}
	err = a.e.SavePolicy()
	if err != nil {
		return errors.WithMessage(err, "保存权限失败")
	}
	return err
}

func (a *authRepo) HasRoleForUser(sys string, uid, role string) (bool, error) {
	ok, err := a.e.HasRoleForUser(prefix(UnityPrefix, uid), prefix(RolePrefix, role), sys)
	if err != nil {
		return false, errors.WithMessage(err, "获取用户角色失败")
	}
	return ok, nil
}

func (a *authRepo) GetUsersForRole(sys, role string) []string {
	us := a.e.GetUsersForRoleInDomain(sys, prefix(RolePrefix, role))
	for i, u := range us {
		us[i] = cleanPrefix(u)
	}
	return us
}

func (a *authRepo) E() *casbin.Enforcer {
	return a.e
}

//go:embed rbac_model.conf
var casbinModel string

func NewAuthRepo(c *conf.Data, cs *conf.Server, logger log.Logger) (biz.AuthRepo, error) {
	pgConf, err := pg.ParseURL(c.Pg.Dsn)
	if err != nil {
		return nil, errors.WithMessage(err, "数据库DSN错误")
	}
	db := pg.Connect(pgConf)
	return newAuthRepo(db, cs.RootAppId, log.NewHelper(logger))
}

func newAuthRepo(db *pg.DB, rootAppId string, logger *log.Helper) (*authRepo, error) {
	a, err := pgadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}
	// casbin RBAC
	m, err := model.NewModelFromString(casbinModel)
	if err != nil {
		return nil, err
	}
	enforcer, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, err
	}
	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, err
	}
	enforcer.AddNamedMatchingFunc("g", "KeyMatch", func(arg1, arg2 string) bool {
		logger.Debugf("keymatch [%v]\t%s == %s", util.KeyMatch(arg1, arg2), arg1, arg2)
		return util.KeyMatch(arg1, arg2)
	})
	enforcer.AddNamedDomainMatchingFunc("g", "KeyMatch", func() defaultrolemanager.MatchingFunc {
		return func(arg1, arg2 string) bool {
			logger.Debugf("compare domain [%v]\t%s vs %s", AsteriskMatch(arg1, arg2), arg1, arg2)
			return AsteriskMatch(arg1, arg2)
		}
	}())
	_, err = enforcer.AddPolicies([][]string{
		[]string{prefix(RolePrefix, "sa"), rootAppId, "*", string(biz.ActCreate)},
		[]string{prefix(RolePrefix, "sa"), rootAppId, "*", string(biz.ActModify)},
		[]string{prefix(RolePrefix, "sa"), rootAppId, "*", string(biz.ActDelete)},
		[]string{prefix(RolePrefix, "sa"), rootAppId, "*", string(biz.ActRead)},
	})

	if err != nil {
		return nil, err
	}
	_, err = enforcer.AddRoleForUserInDomain(fmt.Sprintf("%slocal:root", UnityPrefix), prefix(RolePrefix, "sa"), "*")
	if err != nil {
		return nil, err
	}
	err = enforcer.SavePolicy()
	if err != nil {
		return nil, err
	}
	ar := &authRepo{e: enforcer, logger: logger, rootAppId: rootAppId}
	return ar, nil
}

func AsteriskMatch(key1 string, key2 string) bool {
	return key2 == "*" || key1 == key2
}
func prefix(p, s string) string {
	return p + s
}

func cleanPrefix(s string) string {
	if len(s) < 2 {
		return s
	}
	switch s[0:2] {
	case UnityPrefix, RolePrefix:
		return s[2:]
	default:
		return s
	}
}
