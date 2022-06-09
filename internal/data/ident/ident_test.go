package ident

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/mock/gomock"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/dbmodel"
	"github.com/lunzi/aacs/internal/data/dbtestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepare(t *testing.T) (*biztest.MockThirdPartyRepo, *biztest.MockAccountsRepo, *biztest.MockOpenIDSet, *biztest.MockOpenIDProvider, log.Logger) {
	ctl := gomock.NewController(t)
	tpRepo := biztest.NewMockThirdPartyRepo(ctl)
	accRepo := biztest.NewMockAccountsRepo(ctl)
	openIdRepo := biztest.NewMockOpenIDSet(ctl)
	openIdProviderRepo := biztest.NewMockOpenIDProvider(ctl)
	logger := log.NewStdLogger(os.Stdout)
	return tpRepo, accRepo, openIdRepo, openIdProviderRepo, logger
}
func TestNewIdentRepo(t *testing.T) {

	t.Run("GrantToken", func(t *testing.T) {
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)
		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)
		tpRepo.EXPECT().GetInfo(gomock.Any(), "app").Return(biz.ThirdPartyInfo{
			Name:              "app",
			SecretKey:         "1, 2, 3, 4",
			CallbackUrl:       "http://url.url",
			KeyValidityPeriod: 3600,
		}, nil)
		token, expiredAt, err := repo.GrantToken(context.Background(), "app", "uid")
		require.NoError(t, err)
		t.Log(token, expiredAt)
	})
	t.Run("VerifyToken", func(t *testing.T) {
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)
		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)
		accRepo.EXPECT().GetByID(gomock.Any(), "uid").Return(biz.Account{
			Id:          "uid",
			DisplayName: "UUIIDD",
			Email:       "email@email.com",
			PhoneNo:     "1782910123789",
			Retired:     false,
		}, nil)
		tpRepo.EXPECT().GetInfo(gomock.Any(), "app").Times(2).Return(biz.ThirdPartyInfo{
			Name:              "app",
			SecretKey:         "1, 2, 3, 4",
			CallbackUrl:       "http://url.url",
			KeyValidityPeriod: 3600,
		}, nil)
		token, _, err := repo.GrantToken(context.Background(), "app", "uid")
		require.NoError(t, err)
		s, err := repo.VerifyToken(context.Background(), token)
		require.NoError(t, err)
		t.Log(s)
	})
	t.Run("VerifyAppToken", func(t *testing.T) {
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)
		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)
		tpRepo.EXPECT().GetInfo(gomock.Any(), "app").Times(2).Return(biz.ThirdPartyInfo{
			Name:              "app",
			SecretKey:         "1, 2, 3, 4",
			CallbackUrl:       "http://url.url",
			KeyValidityPeriod: 3600,
		}, nil)
		token, _, err := repo.GrantToken(context.Background(), "app", "")
		require.NoError(t, err)
		s, err := repo.VerifyToken(context.Background(), token)
		require.NoError(t, err)
		t.Log(s)
	})
	t.Run("basic", func(t *testing.T) {
		ctx := context.Background()
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)
		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)
		_, err := db.NewInsert().Model(&dbmodel.Resource{
			Name:      "ldap",
			IsPrimary: false,
			Provider:  "local",
		}).Exec(ctx)
		require.NoError(t, err)
		//openIdRepo.EXPECT().Get(gomock.Any(), "ldap").Return(openIdProviderRepo, true, nil)
		openIdRepo.EXPECT().BasicAuth(gomock.Any(), "ldap", "uid", "pwd").Return(biz.Sub{
			UID:         "tuid",
			DisplayName: "tdn",
			Email:       "temail",
			PhoneNo:     "tphone",
			Source:      "ldap",
			Retired:     false,
		}, nil)
		accRepo.EXPECT().Add(gomock.Any(), gomock.Any(), true).Return(nil)
		accRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(biz.Account{
			Id:          "tuid",
			DisplayName: "tdn",
			Email:       "temail",
			PhoneNo:     "tphone",
			Retired:     false,
		}, nil)
		_, err = repo.Basic(context.Background(), "ldap", "app", "uid", "pwd")
		require.NoError(t, err)

		//u := Unity{}
		//err = db.NewSelect().Model(&u).Where("id=?", "tuid").Scan(context.Background())
		//require.NoError(t, err)
		//assert.Equal(t, u.Id, "tuid")
	})
	t.Run("account_relation", func(t *testing.T) {
		ctx := context.Background()
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)

		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)

		_, err := db.NewInsert().Model(&dbmodel.Resource{
			Name:      "ldap",
			IsPrimary: false,
			Provider:  "local",
		}).Exec(ctx)
		require.NoError(t, err)

		openIdRepo.EXPECT().BasicAuth(gomock.Any(), "ldap", "uid", "pwd").Times(2).Return(biz.Sub{
			UID:         "tuid",
			DisplayName: "tdn",
			Email:       "temail",
			PhoneNo:     "tphone",
			Source:      "ldap",
			Retired:     false,
		}, nil)
		accRepo.EXPECT().Add(gomock.Any(), gomock.Any(), true).Return(nil)
		accRepo.EXPECT().GetByID(gomock.Any(), "tuid").Return(biz.Account{
			Id:          "tuid",
			DisplayName: "tdn",
			Email:       "temail",
			PhoneNo:     "tphone",
			Retired:     false,
		}, nil)
		accRepo.EXPECT().GetByID(gomock.Any(), "replaced").Return(biz.Account{
			Id:          "replaced",
			DisplayName: "replaced",
			Email:       "temail",
			PhoneNo:     "tphone",
			Retired:     false,
		}, nil)
		_, err = repo.Basic(context.Background(), "ldap", "app", "uid", "pwd")
		require.NoError(t, err)
		_, err = db.NewUpdate().Model(&dbmodel.AccountRelation{}).
			Where("ident_source=?", "ldap").Where("ident_id=?", "uid").
			Set("unity_id=?", "replaced").
			Exec(ctx)
		require.NoError(t, err)

		r, err := repo.Basic(context.Background(), "ldap", "app", "uid", "pwd")
		require.NoError(t, err)
		assert.Equal(t, "replaced", r.UID)
		assert.Equal(t, "replaced", r.DisplayName)
	})
	t.Run("retired", func(t *testing.T) {
		ctx := context.Background()
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)
		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)
		//openIdRepo.EXPECT().Get(gomock.Any(), "ldap").Return(openIdProviderRepo, true, nil)
		_, err := db.NewInsert().Model(&dbmodel.Resource{
			Name:      "ldap",
			IsPrimary: false,
			Provider:  "local",
		}).Exec(ctx)
		require.NoError(t, err)
		openIdRepo.EXPECT().BasicAuth(gomock.Any(), "ldap", "uid", "pwd").Return(biz.Sub{
			UID:         "tuid",
			DisplayName: "tdn",
			Email:       "temail",
			PhoneNo:     "tphone",
			Source:      "ldap",
			Retired:     false,
		}, nil)
		accRepo.EXPECT().Add(gomock.Any(), gomock.Any(), true).Return(nil)
		accRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(biz.Account{
			Id:          "tuid",
			DisplayName: "tdn",
			Email:       "temail",
			PhoneNo:     "tphone",
			Retired:     true,
		}, nil)
		_, err = repo.Basic(context.Background(), "ldap", "app", "uid", "pwd")
		require.Error(t, err)
		assert.ErrorContains(t, err, "已离职")
	})
	t.Run("GrantToken", func(t *testing.T) {
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)
		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)
		accRepo.EXPECT().GetByID(gomock.Any(), "uid").Return(biz.Account{
			Id:          "uid",
			DisplayName: "UUIIDD",
			Email:       "email@email.com",
			PhoneNo:     "1782910123789",
			Retired:     false,
		}, nil)
		tpRepo.EXPECT().GetInfo(gomock.Any(), "app").Times(2).Return(biz.ThirdPartyInfo{
			Name:              "app",
			SecretKey:         "1, 2, 3, 4",
			CallbackUrl:       "http://url.url",
			KeyValidityPeriod: 3600,
		}, nil)
		token, expiredAt, err := repo.GrantTokenWithPeriod(context.Background(), "app", "uid", 60*time.Second)
		require.NoError(t, err)
		assert.Condition(t, func() (success bool) {
			return time.Now().Before(expiredAt)
		})
		s, err := repo.VerifyToken(context.Background(), token)
		require.NoError(t, err)
		t.Log(s)
	})

	t.Run("GetUIDByRelation", func(t *testing.T) {
		ctx := context.Background()
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)
		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)
		_, err := repo.GetUIDByRelation(context.Background(), "app", "uid")
		require.Error(t, err)

		ac := dbmodel.Account{
			ID:          "wecom:id",
			DisplayName: "dnn",
			Email:       "emm",
			PhoneNo:     "pnn",
			Retired:     true,
		}
		_, err = db.NewInsert().Model(&ac).Exec(ctx)
		require.NoError(t, err)
		_, err = db.NewInsert().Model(&dbmodel.AccountRelation{
			IdentSource: "wecom",
			IdentID:     "id",
			UnityID:     "wecom:id",
		}).Exec(ctx)
		require.NoError(t, err)

		sub, err := repo.GetUIDByRelation(context.Background(), "wecom", "id")
		require.NoError(t, err)
		assert.Equal(t, "wecom:id", sub.UID)
		assert.Equal(t, "dnn", sub.DisplayName)
		assert.Empty(t, sub.Source)
	})

	t.Run("SaveRelation", func(t *testing.T) {
		ctx := context.Background()
		tpRepo, accRepo, openIdRepo, _, logger := prepare(t)
		db := dbtestutils.GetRandomDB(t)
		repo := NewIdentRepo(tpRepo, accRepo, openIdRepo, db, &conf.Server{RootAppId: "aacs"}, logger)
		err := repo.SaveRelation(ctx, "this_is_uid", "ident_id", "ident_source")
		require.ErrorContains(t, err, "不存在的")

		_, err = db.NewInsert().Model(&dbmodel.Resource{
			Name:      "ident_name",
			IsPrimary: false,
			Provider:  "local",
		}).Exec(ctx)
		require.NoError(t, err)
		err = repo.SaveRelation(ctx, "this_is_uid", "ident_id", "ident_name")
		require.NoError(t, err)
	})
}
