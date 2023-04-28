package geek_web

import (
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewHTTPServer()
	v1 := s.Group("/v1")
	v1.GET("/user", func(ctx *Context) {
		_ = ctx.JSON(http.StatusOK, H{
			"code": 200,
			"msg":  "请求成功" + ctx.Pattern,
		})
	})
	v1.GET("/user/login", func(ctx *Context) {
		_ = ctx.JSON(http.StatusOK, H{
			"code": 200,
			"msg":  "请求成功" + ctx.Pattern,
		})
	})
	v1.GET("/assets/*filepath", func(ctx *Context) {
		filePath, _ := ctx.Param("filepath")
		_ = ctx.JSON(http.StatusOK, H{
			"code": 200,
			"msg":  "请求成功" + ctx.Pattern,
			"info": "你是想访问我的这个文件吗？【" + filePath + "]",
		})
	})
	v1.GET("/user/:id/:action", func(ctx *Context) {
		id, _ := ctx.Param("id")
		action, _ := ctx.Param("action")
		_ = ctx.JSON(http.StatusOK, H{
			"code":   200,
			"msg":    "请求成功" + ctx.Pattern,
			"id":     id,
			"action": action,
		})
	})
	_ = s.Start(":8080")
}
