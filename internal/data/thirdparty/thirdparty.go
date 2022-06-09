package thirdparty

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func NewThirdPartyRepo(logger log.Logger, db *bun.DB, auth biz.AuthRepo) biz.ThirdPartyRepo {
	_, err := db.NewCreateTable().Model(&Model{}).IfNotExists().Exec(context.Background())
	if err != nil {
		panic(errors.WithMessage(err, "初始化第三方应用库失败"))
	}
	tp := &thirdParty{
		log:   log.NewHelper(logger),
		db:    db,
		auth:  auth,
		tpMap: make(map[string]biz.ThirdPartyInfo),
	}

	return tp
}

type thirdParty struct {
	log   *log.Helper
	db    *bun.DB
	auth  biz.AuthRepo
	tpMap map[string]biz.ThirdPartyInfo
	mu    sync.Mutex
}

func (t *thirdParty) VerifyThirdParty(ctx context.Context, appName string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *thirdParty) GetInfo(ctx context.Context, appId string) (biz.ThirdPartyInfo, error) {
	if s, ok := t.tpMap[appId]; ok {
		return s, nil
	}
	m := Model{}
	err := t.db.NewSelect().Model(&m).Where("id=?", appId).Scan(ctx)
	if errors.Is(err, pg.ErrNoRows) {
		return biz.ThirdPartyInfo{}, errors.New("未找到对应的应用")
	}
	if err != nil {
		return biz.ThirdPartyInfo{}, errors.WithMessage(err, "获取应用失败")
	}
	return biz.ThirdPartyInfo{
		Id:                m.ID,
		Name:              m.AppName,
		SecretKey:         m.Secret,
		CallbackUrl:       m.CallbackUrl,
		KeyValidityPeriod: m.KeyValidityPeriod,
		AutoLogin:         m.AutoLogin,
		DevMode:           m.DevMode,
	}, nil
}

func (t *thirdParty) ListAll(ctx context.Context) ([]biz.ThirdPartyInfo, error) {
	var l []Model
	err := t.db.NewSelect().Model(&l).Scan(ctx)
	if err != nil {
		return nil, err
	}
	r := make([]biz.ThirdPartyInfo, len(l))
	for i, m := range l {
		r[i] = biz.ThirdPartyInfo{
			Id:                m.ID,
			Name:              m.AppName,
			SecretKey:         m.Secret,
			CallbackUrl:       m.CallbackUrl,
			KeyValidityPeriod: m.KeyValidityPeriod,
			AutoLogin:         m.AutoLogin,
			DevMode:           m.DevMode,
		}
	}
	t.mu.Lock()
	for i := range r {
		t.tpMap[r[i].Id] = r[i]
	}
	t.mu.Unlock()
	return r, nil
}

func (t *thirdParty) Add(ctx context.Context, appId string, appName string, owner string, callbackUrl string, autoLogin bool) (biz.ThirdPartyInfo, error) {
	if appId == "" {
		appId = uuid.New().String()
	}
	m := Model{
		ID:                appId,
		AppName:           appName,
		Secret:            uuid.New().String(),
		CallbackUrl:       callbackUrl,
		KeyValidityPeriod: 3600 * 7 * 24,
		AutoLogin:         autoLogin,
	}
	_, err := t.db.NewInsert().Model(&m).Exec(ctx)
	if err != nil {
		return biz.ThirdPartyInfo{}, err
	}
	// 添加用户权限
	err = t.auth.AddAdmin(m.ID, owner)
	if err != nil {
		return biz.ThirdPartyInfo{}, errors.WithMessage(err, "绑定管理员权限失败")
	}
	return biz.ThirdPartyInfo{
		Id:                m.ID,
		Name:              m.AppName,
		SecretKey:         m.Secret,
		CallbackUrl:       m.CallbackUrl,
		KeyValidityPeriod: m.KeyValidityPeriod,
		AutoLogin:         m.AutoLogin,
		DevMode:           m.DevMode,
	}, nil
}
