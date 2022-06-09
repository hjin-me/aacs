package pages

import (
	"context"
	"embed"
	"html/template"
	"io"
	rawHttp "net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lunzi/aacs/api/apierr"
	"github.com/pkg/errors"
)

type PageHandler func(ctx context.Context, r *rawHttp.Request, w rawHttp.ResponseWriter) (err error)

type Pages struct {
	pageDebug bool
}

func (p *Pages) WrapHandler(fn PageHandler) http.HandlerFunc {
	return func(c http.Context) (err error) {
		h := c.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			// 注册登陆的回调验证
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, apierr.ErrorNotFound("页面不存在")
			}
			if tr.Kind() != transport.KindHTTP {
				return nil, apierr.ErrorNotFound("协议不支持")
			}
			ht, ok := tr.(http.Transporter)
			if !ok {
				return nil, apierr.ErrorNotFound("协议异常")
			}
			ctx = context.WithValue(ctx, &pageDebug{}, p.pageDebug)
			err := fn(ctx, ht.Request(), c.Response())
			if err != nil {
				return nil, ErrPageBuilder(err)(ctx, ht.Request(), c.Response())
			}
			return nil, nil
		})
		_, err = h(c, nil)
		return err
	}
}

//go:embed *.gohtml
var tplFS embed.FS
var cwd string

type pageDebug struct {
}

func Render(ctx context.Context, tplName string, w io.Writer, data interface{}) error {
	pageDebug, _ := ctx.Value(&pageDebug{}).(bool)

	tpl := template.New("base")
	if pageDebug {
		b, err := os.ReadFile(filepath.Join(cwd, "base.gohtml"))
		if err != nil {
			return errors.WithMessage(err, "没有找到模板 base")
		}
		_, err = tpl.Parse(string(b))
		if err != nil {
			return errors.WithMessagef(err, "编译模板失败 base")
		}
		tpl = tpl.New(tplName)
		b, err = os.ReadFile(filepath.Join(cwd, tplName+".gohtml"))
		if err != nil {
			return errors.WithMessagef(err, "没有找到模板 %s", tplName)
		}
		_, err = tpl.Parse(string(b))
		if err != nil {
			return errors.WithMessagef(err, "编译模板失败 %s", tplName)
		}
	} else {
		b, err := tplFS.ReadFile("base.gohtml")
		if err != nil {
			return errors.WithMessage(err, "没有找到模板 base")
		}
		_, err = tpl.Parse(string(b))
		if err != nil {
			return errors.WithMessagef(err, "编译模板失败 base")
		}
		tpl = tpl.New(tplName)
		b, err = tplFS.ReadFile(tplName + ".gohtml")
		if err != nil {
			return errors.WithMessagef(err, "没有找到模板 %s", tplName)
		}
		_, err = tpl.Parse(string(b))
		if err != nil {
			return errors.WithMessagef(err, "编译模板失败 %s", tplName)
		}
	}

	return tpl.ExecuteTemplate(w, "base", data)
}

func init() {
	_, file, _, _ := runtime.Caller(0)
	cwd = filepath.Dir(file)
}

//func StaticsServer(debug bool) http.Handler {
//	if debug {
//		return http.FileServer(http.Dir(cwd))
//	} else {
//		return http.FileServer(http.FS(tplFS))
//	}
//}
