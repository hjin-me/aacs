package dbtestutils

import (
	"testing"
)

// 测试是否能够成功创建新的数据表
func TestGetTestDB(t *testing.T) {
	db := GetRandomDB(t)
	_ = db
}
