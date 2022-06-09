package assets

import (
	"embed"
	rawHttp "net/http"
	"path/filepath"
	"runtime"
)

//go:embed statics/*
var staticsFS embed.FS
var cwd string

func StaticsServer(debug bool) rawHttp.Handler {
	if debug {
		return rawHttp.FileServer(rawHttp.Dir(cwd))
	} else {
		return rawHttp.FileServer(rawHttp.FS(staticsFS))
	}
}
func init() {
	_, file, _, _ := runtime.Caller(0)
	cwd = filepath.Dir(file)
}
