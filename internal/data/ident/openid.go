package ident

import (
	"context"
	"database/sql"
	"strings"
	"time"

	openidproviderv1 "github.com/lunzi/aacs/api/openidprovider/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/dbmodel"
	"github.com/lunzi/aacs/internal/data/ident/localp"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func NewOpenIdSet(ctx context.Context, db *bun.DB, c *conf.Data, localP *localp.ProviderIns) biz.OpenIDSet {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	s := &openIdSet{db: db}
	_, err := db.NewCreateTable().Model(&dbmodel.Resource{}).IfNotExists().Exec(ctx)
	if err != nil {
		panic(errors.WithMessage(err, "初始化注入器失败"))
	}
	count, err := db.NewSelect().Model(&dbmodel.Resource{}).Where("is_primary = ?", true).Count(ctx)
	if err != nil {
		panic(errors.WithMessage(err, "统计主账号系统失败"))
	}
	if count > 1 {
		panic(errors.New("数据表 openid_idents 存在脏数据，请研发检查"))
	}
	rs := []dbmodel.Resource{
		{Name: "local", IsPrimary: false, Provider: "local"},
	}
	_, err = db.NewInsert().Model(&rs).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		panic(errors.WithMessage(err, "初始化注入器失败"))
	}

	for _, ident := range c.Idents {
		conn, err := grpc.DialContext(ctx,
			ident.Host,
			grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		)
		if err != nil {
			panic(errors.WithMessagef(err, "连接失败 %s", ident.Host))
		}
		err = s.Register(ctx, openidproviderv1.NewOpenIDProviderClient(conn))
		if err != nil {
			panic(errors.WithMessagef(err, "连接失败 %s", ident.Host))
		}

	}
	err = s.Register(ctx, localP)
	if err != nil {
		panic(errors.WithMessage(err, "初始化本地账号库失败"))
	}
	return s
}

type openIdSet struct {
	providerMap map[string]openidproviderv1.OpenIDProviderClient
	db          *bun.DB
}

func (o *openIdSet) Register(ctx context.Context, p openidproviderv1.OpenIDProviderClient) error {
	if o.providerMap == nil {
		o.providerMap = make(map[string]openidproviderv1.OpenIDProviderClient)
	}
	res, err := p.Name(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	o.providerMap[res.Name] = p
	return nil
}

func (o *openIdSet) Get(ctx context.Context, name string) (openidproviderv1.OpenIDProviderClient, bool, error) {
	// 先从数据库里面找到对应的认证途径
	r := dbmodel.Resource{}
	err := o.db.NewSelect().Model(&r).Where("name = ?", name).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, errors.New("没找到对应认证服务")
	}
	if err != nil {
		return nil, false, errors.WithMessage(err, "获取认证服务失败")
	}

	// 用认证途径的配置信息获取对应的认证器
	if p, ok := o.providerMap[r.Provider]; ok {
		return p, r.IsPrimary, nil
	}
	return nil, false, errors.New("不存在的OpenID服务")
}

func (o *openIdSet) ParseUID(ctx context.Context, uid string) (ns, id string, err error) {
	if uid == "" {
		return "", "", errors.New("uid 为空")
	}
	o.db.NewSelect()
	part := strings.Split(uid, ":")
	if len(part) == 1 {
		// default
		r := dbmodel.Resource{}
		err := o.db.NewSelect().Model(&r).Where(`is_default = ?`, true).Scan(ctx)
		if err != nil {
			return "", "", errors.WithMessage(err, "没有默认账号源")
		}
		return r.Name, part[0], nil
	}
	ns = part[0]
	id = part[1]
	return ns, id, nil
}

func (o *openIdSet) buildUID(source, id string, isDefault bool) string {
	if isDefault {
		return id
	}
	return source + ":" + id
}

func (o *openIdSet) BasicAuth(ctx context.Context, name, uid, pwd string) (biz.Sub, error) {
	op, isDefault, err := o.Get(ctx, name)
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "登陆服务内部错误")
	}
	res, err := op.BasicAuth(ctx, &openidproviderv1.BasicAuthReq{
		Uid: uid,
		Pwd: pwd,
	})
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "登陆失败")
	}
	res.Sub.Uid = o.buildUID(name, res.Sub.Uid, isDefault)
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "BasicAuth 失败")
	}
	return biz.Sub{
		UID:         res.Sub.Uid,
		DisplayName: res.Sub.DisplayName,
		Email:       res.Sub.Email,
		PhoneNo:     res.Sub.PhoneNo,
		Source:      res.Sub.Source,
		App:         res.Sub.App,
		Retired:     res.Sub.Retired,
		Gender:      res.Sub.Gender,
	}, nil
}

func (o *openIdSet) SearchUid(ctx context.Context, name, uid string) (biz.Sub, error) {
	p, isDefault, err := o.Get(ctx, name)
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "没有账号源")
	}
	res, err := p.SearchUid(ctx, &openidproviderv1.SearchUidReq{
		Uid: uid,
	})
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "查找用户失败")
	}
	res.Sub.Source = name
	res.Sub.Uid = o.buildUID(name, res.Sub.Uid, isDefault)
	return biz.Sub{
		UID:         res.Sub.Uid,
		DisplayName: res.Sub.DisplayName,
		Email:       res.Sub.Email,
		PhoneNo:     res.Sub.PhoneNo,
		Source:      res.Sub.Source,
		App:         res.Sub.App,
		Retired:     res.Sub.Retired,
		Gender:      res.Sub.Gender,
	}, nil
}

func (o *openIdSet) TokenAuth(ctx context.Context, name, token string) (biz.Sub, string, error) {
	op, isDefault, err := o.Get(ctx, name)
	if err != nil {
		return biz.Sub{}, "", errors.WithMessage(err, "登陆服务内部错误")
	}
	res, err := op.TokenAuth(ctx, &openidproviderv1.TokenAuthReq{Token: token})
	if err != nil {
		return biz.Sub{}, "", errors.WithMessage(err, "登陆失败")
	}
	uid := res.Uid
	sub := res.Sub
	sub.Uid = o.buildUID(name, uid, isDefault)
	if err != nil {
		return biz.Sub{}, "", errors.WithMessage(err, "BasicAuth 失败")
	}
	return biz.Sub{
		UID:         sub.Uid,
		DisplayName: sub.DisplayName,
		Email:       sub.Email,
		PhoneNo:     sub.PhoneNo,
		Source:      sub.Source,
		App:         sub.App,
		Retired:     sub.Retired,
		Gender:      sub.Gender,
	}, uid, nil
}
