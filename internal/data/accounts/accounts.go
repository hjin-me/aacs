package accounts

import (
	"context"
	"database/sql"
	"time"

	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/data/dbmodel"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func NewAccountsRepo(db *bun.DB, opSet biz.OpenIDSet) biz.AccountsRepo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := db.NewCreateTable().Model(&dbmodel.Account{}).IfNotExists().Exec(ctx)
	if err != nil {
		panic(errors.WithMessage(err, "NewAccountsRepo"))
	}
	return &accountsRepo{db: db, opSet: opSet}
}

type accountsRepo struct {
	db    *bun.DB
	opSet biz.OpenIDSet
}

func (a *accountsRepo) Add(ctx context.Context, sub biz.Account, ignoreConflict bool) error {
	// 同步至本地数据库
	u := dbmodel.Account{
		ID:          sub.Id,
		DisplayName: sub.DisplayName,
		Email:       sub.Email,
		PhoneNo:     sub.PhoneNo,
	}
	q := a.db.NewInsert().Model(&u)
	if ignoreConflict {
		q = q.On("CONFLICT (id) DO UPDATE").
			Set("display_name=?", u.DisplayName).
			Set("email=?", u.Email).
			Set("phone_no=?", u.PhoneNo)
	}
	_, err := q.Exec(ctx)
	if err != nil {
		return errors.WithMessage(err, "登陆服务内部错误")
	}

	return nil
}

func (a *accountsRepo) Save(ctx context.Context, s biz.Account) error {
	u := dbmodel.Account{
		ID:          s.Id,
		DisplayName: s.DisplayName,
		Email:       s.Email,
		PhoneNo:     s.PhoneNo,
		Retired:     s.Retired,
	}
	_, err := a.db.NewInsert().Model(&u).
		On("CONFLICT (id) DO UPDATE").
		Set("display_name=?", u.DisplayName).
		Set("email=?", u.Email).
		Set("phone_no=?", u.PhoneNo).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *accountsRepo) AllSubject(ctx context.Context) ([]biz.Account, error) {
	var us []dbmodel.Account

	err := a.db.NewSelect().Model(&us).
		Relation("AccountRelations").
		Limit(1000).Scan(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "获取用户列表失败")
	}
	bs := make([]biz.Account, len(us))
	for i2, u := range us {
		ba := biz.Account{
			Id:          u.ID,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			PhoneNo:     u.PhoneNo,
			Retired:     u.Retired,
		}
		ba.RelatedIdents = make([]biz.Ident, len(u.AccountRelations))
		for i, relation := range u.AccountRelations {
			ba.RelatedIdents[i] = biz.Ident{
				Source: relation.IdentSource,
				Id:     relation.IdentID,
			}
		}

		bs[i2] = ba
	}
	return bs, nil
}

func (a *accountsRepo) GetByID(ctx context.Context, id string) (biz.Account, error) {
	u := dbmodel.Account{}
	err := a.db.NewSelect().Model(&u).Where("id=?", id).
		Where("deleted_at is null").
		Relation("AccountRelations").
		Limit(1).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return biz.Account{}, errors.Errorf("用户名不存在 %s", id)
	}
	if err != nil {
		return biz.Account{}, err
	}
	ba := biz.Account{
		Id:          u.ID,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		PhoneNo:     u.PhoneNo,
		Retired:     u.Retired,
	}
	ba.RelatedIdents = make([]biz.Ident, len(u.AccountRelations))
	for i, relation := range u.AccountRelations {
		ba.RelatedIdents[i] = biz.Ident{
			Source: relation.IdentSource,
			Id:     relation.IdentID,
		}
	}
	return ba, nil
}

// SyncWecom 该接口只同步更新未设置企业微信的账号
func (a *accountsRepo) SyncWecom(ctx context.Context) error {
	//all, err := a.wx.GetUsers(ctx)
	//if err != nil {
	//	return errors.WithMessage(err, "获取用户列表失败")
	//}
	//var names []string
	//for _, employee := range all {
	//	names = append(names, employee.Name)
	//}
	//var us []dbmodel.Account
	//err = a.db.NewSelect().Model(&us).
	//	Where("display_name=ANY(?)", pgdialect.Array(names)).
	//	Relation("AccountRelations").Scan(ctx)
	//if errors.Is(err, sql.ErrNoRows) {
	//	return errors.New("没有找到匹配的用户名")
	//}
	//if err != nil {
	//	return errors.WithMessage(err, "获取用户列表失败")
	//}
	//var ars []dbmodel.AccountRelation
	//for _, u := range us {
	//	if hasWecom(u) {
	//		continue
	//	}
	//	for _, wu := range all {
	//		if wu.Name == u.DisplayName {
	//			ars = append(ars, dbmodel.AccountRelation{
	//				IdentSource: "wecom",
	//				IdentID:     wu.UserId,
	//				UnityID:     u.ID,
	//			})
	//			break
	//		}
	//	}
	//}
	//if len(ars) == 0 {
	//	return errors.New("没有更新任何账号数据")
	//}
	//_, err = a.db.NewInsert().Model(&ars).On("CONFLICT DO NOTHING").Exec(ctx)
	//if err != nil {
	//	return errors.WithMessage(err, "更新企业微信信息时失败")
	//}
	//return nil
	return nil
}

func (a *accountsRepo) ImportAccount(ctx context.Context, identSource string, uid string) (biz.Account, error) {
	sub, err := a.opSet.SearchUid(ctx, identSource, uid)
	if err != nil {
		return biz.Account{}, errors.WithMessage(err, "未能找到对应用户")
	}
	ac := biz.Account{
		Id:          sub.UID,
		DisplayName: sub.DisplayName,
		Email:       sub.Email,
		PhoneNo:     sub.PhoneNo,
		Retired:     sub.Retired,
		AllowedApps: nil,
	}
	err = a.Add(ctx, ac, true)
	if err != nil {
		return biz.Account{}, errors.WithMessage(err, "导入用户失败")
	}
	return ac, nil
}
