package accounts

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lunzi/aacs/internal/biz"
	"github.com/lunzi/aacs/internal/biz/biztest"
	"github.com/lunzi/aacs/internal/data/dbmodel"
	"github.com/lunzi/aacs/internal/data/dbtestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountsRepo(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		var err error
		ctx := context.Background()
		db := dbtestutils.GetRandomDB(t)
		_, err = db.NewCreateTable().Model(&dbmodel.AccountRelation{}).IfNotExists().Exec(context.Background())
		require.NoError(t, err)
		_, err = db.NewInsert().Model(&dbmodel.AccountRelation{
			IdentSource: "sss",
			IdentID:     "xxx",
			UnityID:     "aaaa",
		}).Exec(ctx)
		require.NoError(t, err)

		ctl := gomock.NewController(t)
		opSet := biztest.NewMockOpenIDSet(ctl)
		repo := NewAccountsRepo(db, opSet)
		err = repo.Add(ctx, biz.Account{
			Id:          "aaaa",
			DisplayName: "bbbb",
			Email:       "ccc@ccc.com",
			PhoneNo:     "ddddd",
		}, true)
		require.NoError(t, err)
		err = repo.Add(ctx, biz.Account{
			Id:          "aaaa",
			DisplayName: "dn",
			Email:       "dn@ccc.com",
			PhoneNo:     "dn",
		}, true)
		require.NoError(t, err)
		s, err := repo.GetByID(ctx, "aaaa")
		require.NoError(t, err)
		assert.Equal(t, "dn", s.DisplayName)
		assert.Equal(t, "xxx", s.RelatedIdents[0].Id)

		err = repo.Add(ctx, biz.Account{
			Id:          "aaaa",
			DisplayName: "bbbb",
			Email:       "ccc@ccc.com",
			PhoneNo:     "ddddd",
		}, false)
		require.Error(t, err)

		_, err = repo.AllSubject(ctx)
		require.NoError(t, err)
	})
}
