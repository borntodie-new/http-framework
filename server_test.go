package geek_web

import (
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
	_ = s.Start(":8080")
}
