package localp

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	openidproviderv1 "github.com/lunzi/aacs/api/openidprovider/v1"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ openidproviderv1.OpenIDProviderClient = (*ProviderIns)(nil)

func NewPgProvider(db *bun.DB, l log.Logger) *ProviderIns {
	_, err := db.NewCreateTable().Model(&LocalUser{}).IfNotExists().Exec(context.Background())
	if err != nil {
		panic(errors.WithMessage(err, "初始化本地用户表失败"))
	}
	p := &ProviderIns{db: db, name: "local", logger: log.NewHelper(l)}
	_ = p.Create(context.Background(), "root", "6a7c0235-2c88-4ca1-862c-28921d981c79", "默认账号", "默认邮箱", "没有手机号")
	return p
}

type ProviderIns struct {
	logger *log.Helper
	db     *bun.DB
	name   string
}

func (l *ProviderIns) Name(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*openidproviderv1.NameReply, error) {
	return &openidproviderv1.NameReply{Name: l.name}, nil
}

func (l *ProviderIns) BasicAuth(ctx context.Context, in *openidproviderv1.BasicAuthReq, opts ...grpc.CallOption) (*openidproviderv1.BasicAuthReply, error) {
	res, err := l.basic(ctx, in.Uid, in.Pwd)
	if err != nil {
		return &openidproviderv1.BasicAuthReply{}, err
	}
	return &openidproviderv1.BasicAuthReply{
		Sub: &openidproviderv1.Subject{
			Uid:         res.UID,
			DisplayName: res.DisplayName,
			Email:       res.Email,
			PhoneNo:     res.PhoneNo,
			Source:      res.Source,
			App:         res.App,
			Retired:     res.Retired,
			Gender:      res.Gender,
		},
	}, nil
}

func (l *ProviderIns) TokenAuth(ctx context.Context, in *openidproviderv1.TokenAuthReq, opts ...grpc.CallOption) (*openidproviderv1.TokenAuthReply, error) {
	return &openidproviderv1.TokenAuthReply{}, errors.New("不支持的接口")
}

func (l *ProviderIns) SearchUid(ctx context.Context, in *openidproviderv1.SearchUidReq, opts ...grpc.CallOption) (*openidproviderv1.SearchUidReply, error) {
	res, err := l.searchUid(ctx, in.Uid)
	if err != nil {
		return &openidproviderv1.SearchUidReply{}, err
	}
	return &openidproviderv1.SearchUidReply{
		Sub: &openidproviderv1.Subject{
			Uid:         res.UID,
			DisplayName: res.DisplayName,
			Email:       res.Email,
			PhoneNo:     res.PhoneNo,
			Source:      res.Source,
			App:         res.App,
			Retired:     res.Retired,
			Gender:      res.Gender,
		},
	}, nil
}

func (l *ProviderIns) signPwd(salt, pwd string) string {
	h := sha256.New()
	return hex.EncodeToString(h.Sum([]byte(pwd + salt)))
}
func (l *ProviderIns) basic(ctx context.Context, uid, pwd string) (biz.Sub, error) {
	user := LocalUser{}
	err := l.db.NewSelect().Model(&user).Where("id = ?", uid).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return biz.Sub{}, errors.New("用户信息有误")
	}
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "内部服务错误")
	}
	p := l.signPwd(user.Salt, pwd)
	if p != user.Pwd {
		l.logger.Warnf("用户 %s 密码不匹配 %s", uid, p)
		return biz.Sub{}, errors.New("用户名密码不匹配")
	}

	return biz.Sub{
		UID:         user.ID,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		PhoneNo:     user.PhoneNo,
		Source:      l.name,
		Retired:     user.Retired,
	}, nil
}

func (l *ProviderIns) Create(ctx context.Context, uid, pwd, name, email, phoneNo string) error {
	u := LocalUser{
		ID:          uid,
		Salt:        uuid.New().String(),
		DisplayName: name,
		Email:       email,
		PhoneNo:     phoneNo,
		Retired:     false,
	}
	h := sha256.New()
	p := hex.EncodeToString(h.Sum([]byte(pwd + u.Salt)))
	u.Pwd = p

	_, err := l.db.NewInsert().Model(&u).Exec(ctx)
	if err != nil {
		return errors.WithMessage(err, "创建用户失败")
	}
	return nil
}
func (l *ProviderIns) ResetPwd(ctx context.Context, uid, oldPwd, newPwd string) error {
	u := LocalUser{
		ID: uid,
	}
	err := l.db.NewSelect().Model(&u).Where("id = ?", uid).Scan(ctx)
	if err != nil {
		return errors.WithMessagef(err, "查找用户失败, %s", uid)
	}
	p := l.signPwd(u.Salt, oldPwd)
	if u.Pwd != p {
		return errors.New("密码不匹配")
	}
	newSalt := uuid.New().String()
	np := l.signPwd(newSalt, newPwd)
	_, err = l.db.NewUpdate().Model(&u).
		Set("pwd = ?", np).
		Set("salt = ?", newSalt).
		Where("id = ?", uid).Exec(ctx)
	if err != nil {
		return errors.WithMessage(err, "更新密码失败")
	}
	return nil
}

func (l *ProviderIns) searchUid(ctx context.Context, uid string) (biz.Sub, error) {
	user := LocalUser{}
	err := l.db.NewSelect().Model(&user).Where("id = ?", uid).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return biz.Sub{}, errors.New("用户信息有误")
	}
	if err != nil {
		return biz.Sub{}, errors.WithMessage(err, "内部服务错误")
	}

	return biz.Sub{
		UID:         user.ID,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		PhoneNo:     user.PhoneNo,
		Source:      l.name,
		Retired:     user.Retired,
	}, nil
}
