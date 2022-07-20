package ident

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/dbmodel"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func NewIdentRepo(tp biz.ThirdPartyRepo, acc biz.AccountsRepo, opSet biz.OpenIDSet,
	db *bun.DB,
	cs *conf.Server, logger log.Logger) biz.IdentRepo {
	_, err := db.NewCreateTable().Model(&dbmodel.Account{}).IfNotExists().Exec(context.Background())
	if err != nil {
		panic(errors.WithMessage(err, "NewIdentRepo"))
	}
	_, err = db.NewCreateTable().Model(&dbmodel.Resource{}).IfNotExists().Exec(context.Background())
	if err != nil {
		panic(errors.WithMessage(err, "NewIdentRepo"))
	}
	_, err = db.NewCreateTable().Model(&dbmodel.AccountRelation{}).IfNotExists().Exec(context.Background())
	if err != nil {
		panic(errors.WithMessage(err, "NewIdentRepo"))
	}

	return &identRepo{
		log:       log.NewHelper(logger),
		tp:        tp,
		opSet:     opSet,
		rootAppId: cs.RootAppId,
		acc:       acc,
		db:        db,
	}
}

type identRepo struct {
	log       *log.Helper
	tp        biz.ThirdPartyRepo
	acc       biz.AccountsRepo
	opSet     biz.OpenIDSet
	db        *bun.DB
	rootAppId string
}

func (i *identRepo) Basic(ctx context.Context, source string, app string, uid string, pwd string) (biz.Sub, error) {
	sub, err := i.opSet.BasicAuth(ctx, source, uid, pwd)
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "登陆失败")
	}

	return i.afterAuth(ctx, source, app, uid, sub)
}
func (i *identRepo) TokenAuth(ctx context.Context, source string, app string, token string) (biz.Sub, error) {
	sub, uid, err := i.opSet.TokenAuth(ctx, source, token)
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "登陆失败")
	}
	return i.afterAuth(ctx, source, app, uid, sub)

}
func (i *identRepo) afterAuth(ctx context.Context, source string, app string, uid string, sub biz.Sub) (biz.Sub, error) {
	ar := dbmodel.AccountRelation{}
	exist, err := i.db.NewSelect().Model(&ar).Where("ident_id=?", uid).Where("ident_source=?", source).Exists(ctx)
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "登陆失败")
	}
	if !exist {
		err = i.acc.Add(ctx, biz.Account{
			Id:          sub.UID,
			DisplayName: sub.DisplayName,
			Email:       sub.Email,
			PhoneNo:     sub.PhoneNo,
			Retired:     false,
		}, true)
		if err != nil {
			return biz.Sub{}, errors.WithMessage(err, "添加账户失败")
		}
		err = i.SaveRelation(ctx, sub.UID, uid, source)
		if err != nil {
			return biz.Sub{}, errors.WithMessage(err, "绑定账号关系失败")
		}
	}
	err = i.db.NewSelect().Model(&ar).Where("ident_id=?", uid).Where("ident_source=?", source).Scan(ctx)
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "登陆服务内部错误")
	}
	acc, err := i.acc.GetByID(ctx, ar.UnityID)
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "登陆服务内部错误")
	}
	if acc.Retired {
		return biz.Sub{}, errors.New("该用户已离职，如果需要强制登陆，请联系管理员")
	}

	return biz.Sub{
		UID:         acc.Id,
		DisplayName: acc.DisplayName,
		Email:       acc.Email,
		PhoneNo:     acc.PhoneNo,
		App:         app,
		Retired:     acc.Retired,
	}, nil
}

func (i *identRepo) GrantToken(ctx context.Context, app, uid string) (string, time.Time, error) {
	appInfo, err := i.tp.GetInfo(ctx, app)
	if err != nil {
		return "", time.Time{}, errors.WithMessagef(err, "获取应用信息失败[%s]", app)
	}
	expiredAt := time.Now().Add(time.Duration(appInfo.KeyValidityPeriod) * time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, biz.Claims{
		Issuer:    i.rootAppId,
		Subject:   uid,
		Audience:  jwt.ClaimStrings{app},
		ExpiresAt: jwt.NewNumericDate(expiredAt),
		NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Version:   "v1",
		Domain:    app,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(appInfo.SecretKey))
	if err != nil {
		return "", time.Time{}, errors.WithMessage(err, "token签名失败")
	}
	return tokenString, expiredAt, nil
}
func (i *identRepo) GrantTokenWithPeriod(ctx context.Context, app, uid string, p time.Duration) (string, time.Time, error) {
	appInfo, err := i.tp.GetInfo(ctx, app)
	if err != nil {
		return "", time.Time{}, errors.WithMessagef(err, "获取应用信息失败[%s]", app)
	}
	expiredAt := time.Now().Add(p)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, biz.Claims{
		Issuer:    i.rootAppId,
		Subject:   uid,
		Audience:  jwt.ClaimStrings{app},
		ExpiresAt: jwt.NewNumericDate(expiredAt),
		NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Version:   "v1",
		Domain:    app,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(appInfo.SecretKey))
	if err != nil {
		return "", time.Time{}, errors.WithMessage(err, "token签名失败")
	}
	return tokenString, expiredAt, nil
}

func (i *identRepo) VerifyToken(ctx context.Context, tokenStr string) (biz.Sub, error) {
	if tokenStr == "" {
		return biz.Sub{}, errors.New("没有 token")
	}
	var ok bool
	var claims *biz.Claims
	token, err := jwt.ParseWithClaims(tokenStr, &biz.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok = token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if claims, ok = token.Claims.(*biz.Claims); !ok {
			return nil, fmt.Errorf("unexpected 验证错误")
		}
		if claims.Domain == "" {
			return nil, errors.New("token 没有颁发对象")
		}
		appInfo, err := i.tp.GetInfo(ctx, claims.Domain)
		if err != nil {
			return nil, errors.WithMessage(err, "获取token对应的应用失败")
		}
		return []byte(appInfo.SecretKey), nil
	})

	if err != nil {
		return biz.Sub{}, err
	}
	if !token.Valid {
		return biz.Sub{}, fmt.Errorf("token 不合法")
	}
	if claims, ok = token.Claims.(*biz.Claims); !ok {
		return biz.Sub{}, fmt.Errorf("token 还是不合法")
	}
	if claims.Subject == "" {
		return biz.Sub{
			App: claims.Domain,
		}, nil
	}

	acc, err := i.acc.GetByID(ctx, claims.Subject)
	if err != nil {
		return biz.Sub{}, err
	}
	return biz.Sub{
		UID:         acc.Id,
		DisplayName: acc.DisplayName,
		Email:       acc.Email,
		PhoneNo:     acc.PhoneNo,
		App:         claims.Domain,
		Retired:     acc.Retired,
	}, nil
}

func (i *identRepo) ParseUID(ctx context.Context, uid string) (ns, id string, err error) {
	return i.opSet.ParseUID(ctx, uid)
}
func (i *identRepo) SaveRelation(ctx context.Context, uId, identId, identSource string) error {
	if uId == "" || identId == "" || identSource == "" {
		return errors.Errorf("关系表信息不完整, unityId: `%s`, identId: `%s`, identSource: `%s`", uId, identId, identSource)
	}
	ok, err := i.db.NewSelect().Model(&dbmodel.Resource{}).Where("name=?", identSource).Exists(ctx)
	if err != nil {
		return errors.WithMessage(err, "插入账户关系操作失败")
	}
	if !ok {
		return errors.Errorf("不存在的 ident 源, [%s]", identSource)
	}
	ar := dbmodel.AccountRelation{
		IdentSource: identSource,
		IdentID:     identId,
		UnityID:     uId,
	}
	_, err = i.db.NewInsert().Model(&ar).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		return errors.WithMessage(err, "插入账户关系失败")
	}
	return nil
}

func (i *identRepo) GetUIDByRelation(ctx context.Context, identSource, id string) (biz.Sub, error) {
	ac := dbmodel.Account{}
	err := i.db.NewSelect().Model(&ac).
		//ColumnExpr("accounts.*").
		Join("JOIN account_relations AS a ").
		JoinOn("a.unity_id = account.id").
		JoinOn("a.ident_source=?", identSource).
		JoinOn("a.ident_id=?", id).
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return biz.Sub{}, errors.New("没有找到用户")
	}
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "查找账户关系失败")
	}
	return biz.Sub{
		UID:         ac.ID,
		DisplayName: ac.DisplayName,
		Email:       ac.Email,
		PhoneNo:     ac.PhoneNo,
		Retired:     ac.Retired,
	}, nil
}
