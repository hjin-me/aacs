# 认证系统整体接入流程

1. 在认证系统注册一个应用，并获取ID和密钥
2. 获取认证系统的相关 golang SDK
3. 注册一个回调地址，当认证系统验证用户身份成功后，携带该用户的token访问回调系统
4. 应用通过 token 获取并认证用户信息

# 如何后端如何链接认证系统

访问 aacs GRPC 接口时需要做 token 验证，可以访问 查看你有管理权限的应用 id 和 密钥

```golang
package some

import (
	v1 "github.com/lunzi/aacs/api/identification/v1"
	"github.com/lunzi/aacs/pkgs/connauth"
)

func getIdent() v1.IdentificationClient{
	conn, _ := grpc.Dial("认证服务后端地址",
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(),
		grpc.WithUnaryInterceptor(connauth.Client("应用ID", []byte("应用密钥"))),
	)
	return v1.NewIdentificationClient(conn)
}

```

# 后端如何响应用户登陆成功后的回调

回调地址包含两个参数 `x-aacs-token` token 具体内容、 `x-aacs-expired-at` token 有效期

使用方法

## 纯后端接入

### 收到 token 后的处理
该方案接入后，后端会将 token 写入 cookies 中，后续同域下前后端请求都能包含身份认证信息
```golang
package main
import (
	"github.com/lunzi/aacs/pkgs/authmd"
)
func main() {
	// 配合 kratos 使用
	srv := http.NewServer()
	ident := authmd.NewIdent(getIdent(), "应用ID")

	authmd.NewAuthCallbackServ(srv, &userRepo{u: user}, ident, "/callback", "/", log.NewHelper(logger))
}

```

### 其他请求中 token 数据的传递和解析
```golang
package main
import (

	"github.com/lunzi/aacs/pkgs/authmd"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
)
func NewHTTPServer() {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			// 引入 metadata 解析 header
			metadata.Server(
				metadata.WithPropagatedPrefix("")),
			// 引入 authmd 中间件解析 token	
			authmd.Server(log.NewHelper(logger), ident),
		),
	}

	// 配合 kratos 使用
	srv := http.NewServer()
	
	// 请求处理函数内读取用户信息
	uid, token, _ := authmd.GetSession(ctx)
	uid, err := authmd.GetUID(ctx)
}
```

## 纯前端接入方案

TODO

