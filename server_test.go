package geek_web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewHTTPServer()
	s.GET("/user", func(ctx *Context) {
		ctx.Response.WriteHeader(http.StatusOK)
		_, _ = ctx.Response.Write([]byte("/user"))
	})
	s.GET("/user/login", func(ctx *Context) {
		ctx.Response.WriteHeader(http.StatusOK)
		_, _ = ctx.Response.Write([]byte("/user/login"))
	})
	s.GET("/assets/*filepath", func(ctx *Context) {
		filepath := ctx.Params["filepath"]
		ctx.Response.WriteHeader(http.StatusOK)
		_, _ = ctx.Response.Write([]byte(fmt.Sprintf("你是想找【%s】文件吗？", filepath)))
	})
	s.GET("/user/:id/:action", func(ctx *Context) {
		id := ctx.Params["id"]
		action := ctx.Params["action"]
		ctx.Response.WriteHeader(http.StatusOK)
		_, _ = ctx.Response.Write([]byte(fmt.Sprintf("你是不是想对ID是%s的用户进行%s操作", id, action)))
	})
	_ = s.Start(":8080")
}
