package pfsession

// x
// 这里记录一下老PFSession里面出现过的字段
type x struct {
	Manager struct {
		Cn              string `json:"cn"`
		Id              string `json:"_id"`
		MemberId        string `json:"member_id"`
		Username        string `json:"username"`
		Description     string `json:"description"`
		Displayname     string `json:"displayname"`
		Gidnumber       string `json:"gidnumber"`
		Homedirectory   string `json:"homedirectory"`
		Loginshell      string `json:"loginshell"`
		Mail            string `json:"mail"`
		O               string `json:"o"`
		Sn              string `json:"sn"`
		Telephonenumber string `json:"telephonenumber"`
		Uid             string `json:"uid"`
		Uidnumber       string `json:"uidnumber"`
		Userpassword    string `json:"userpassword"`
		Dn              string `json:"dn"`
		Group           string `json:"group"`
		Email           string `json:"email"`
		Phone           string `json:"phone"`
		Realname        string `json:"realname"`
		Token           string `json:"token"`
		Authenticate    int    `json:"authenticate"`
		DhTest          string `json:"dh_test"`
	} `json:"manager"`
}

var _ *x = nil
