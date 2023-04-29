package geek_web

import (
	"fmt"
	"github.com/borntodie-new/geek-web/middleware/accesslog"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	s := NewHTTPServer()
	// 我们注册一个全局的中间件
	s.Use(func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			ctime := time.Now()
			time.Sleep(time.Microsecond * 2)
			next(ctx)
			fmt.Printf("请求总耗时：%d 毫秒\n", time.Since(ctime).Microseconds())
		}
	})
	// v1和v2两个路由组只有请求方法不一致，其他都是一样的
	v1 := s.Group("/v1")
	builder := accesslog.NewBuilder(nil)
	// 给v1注册日志记录中间件
	v1.Use(builder.Builder())
	{
		v1.GET("/user", func(ctx *Context) {
			ctx.JSON(http.StatusOK, H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"method": ctx.Method,
			})
		})
		v1.GET("/user/login", func(ctx *Context) {
			ctx.JSON(http.StatusOK, H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"method": ctx.Method,
			})
		})
		v1.GET("/assets/*filepath", func(ctx *Context) {
			filePath, _ := ctx.Param("filepath")
			ctx.JSON(http.StatusOK, H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"info":   "你是想访问我的这个文件吗？【" + filePath + "]",
				"method": ctx.Method,
			})
		})
		v1.GET("/user/:id/:action", func(ctx *Context) {
			id, _ := ctx.Param("id")
			action, _ := ctx.Param("action")
			ctx.JSON(http.StatusOK, H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"id":     id,
				"action": action,
				"method": ctx.Method,
			})
		})
	}
	v2 := s.Group("/v2")
	{
		v2.POST("/user", func(ctx *Context) {
			ctx.JSON(http.StatusOK, H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"method": ctx.Method,
			})
		})
		v2.POST("/user/login", func(ctx *Context) {
			ctx.JSON(http.StatusOK, H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"method": ctx.Method,
			})
		})
		v2.POST("/assets/*filepath", func(ctx *Context) {
			filePath, _ := ctx.Param("filepath")
			ctx.JSON(http.StatusOK, H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"info":   "你是想访问我的这个文件吗？【" + filePath + "]",
				"method": ctx.Method,
			})
		})
		v2.POST("/user/:id/:action", func(ctx *Context) {
			id, _ := ctx.Param("id")
			action, _ := ctx.Param("action")
			ctx.JSON(http.StatusOK, H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"id":     id,
				"action": action,
				"method": ctx.Method,
			})
		})
	}
	_ = s.Start(":8080")
}
