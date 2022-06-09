package pfsession

import (
	"context"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/mock/gomock"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/lunzi/aacs/internal/data/myotel/myoteltest"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	ctl := gomock.NewController(t)
	spanRepo := myoteltest.NewMockSpanExporter(ctl)
	ro := NewRedisConf(&conf.Data{Redis: defaultDBConf()})
	p := NewPfSession(context.Background(), ro, spanRepo, log.NewStdLogger(os.Stdout))
	assert.Equal(t, `manager|a:2:{s:8:"username";s:8:"huangjin";s:2:"cn";s:8:"huangjin";}`, p.serialize("huangjin"))

	// aa := "manager|a:24:{s:2:\"cn\";s:8:\"huangjin\";s:3:\"_id\";s:8:\"huangjin\";s:9:\"member_id\";s:8:\"huangjin\";s:8:\"username\";s:8:\"huangjin\";s:11:\"description\";s:8:\"6buE6L+b\";s:11:\"displayname\";s:6:\"\xe9\xbb\x84\xe8\xbf\x9b\";s:9:\"gidnumber\";s:5:\"10005\";s:13:\"homedirectory\";s:14:\"/home/huangjin\";s:10:\"loginshell\";s:9:\"/bin/bash\";s:4:\"mail\";s:17:\"huangjin@antiy.cn\";s:1:\"o\";s:32:\"6a44426d691cd1309029f5e470ee142d\";s:2:\"sn\";s:8:\"huangjin\";s:15:\"telephonenumber\";s:11:\"\";s:3:\"uid\";s:8:\"huangjin\";s:9:\"uidnumber\";s:5:\"10166\";s:12:\"userpassword\";s:29:\"{MD5}/QX1QzNlgMU9YLHYirLsUA==\";s:2:\"dn\";s:1:\"c\";s:5:\"group\";s:5:\"10005\";s:5:\"email\";s:17:\"huangjin@antiy.cn\";s:5:\"phone\";s:11:\"18810014483\";s:8:\"realname\";s:6:\"\xe9\xbb\x84\xe8\xbf\x9b\";s:5:\"token\";s:32:\"6a44426d691cd1309029f5e470ee142d\";s:12:\"authenticate\";i:1;s:7:\"dh_test\";s:4:\"ldap\";}"
}
