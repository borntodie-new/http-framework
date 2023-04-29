package geek_web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewHTTPServer()
	v1 := s.Group("/v1")
	v1.Use(func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			fmt.Println("coming middleware1...")
			next(ctx)
			fmt.Println("outing middleware1...")
		}
	}, func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			fmt.Println("coming middleware2...")
			next(ctx)
			fmt.Println("outing middleware2...")
		}
	}, func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			fmt.Println("coming middleware3...")
			next(ctx)
			fmt.Println("outing middleware3...")
		}
	}, func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			fmt.Println("coming middleware4...")
			next(ctx)
			fmt.Println("outing middleware4...")
		}
	})
	v2 := s.Group("/v2")
	v2.Use(func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {

		}
	}, func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {

		}
	}, func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {

		}
	}, func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {

		}
	})
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
