package test

import (
	"fmt"
	"github.com/borntodie-new/geek-web"
	"github.com/borntodie-new/geek-web/middleware/accesslog"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	tplEngine := geek_web.NewGoTemplateEngine()
	// 初始化server是加上配置
	// 表示需要一个带有TemplateEngine的server
	s := geek_web.NewHTTPServer(geek_web.ServerWithTemplateEngine(tplEngine))
	// 我们注册一个全局的中间件
	s.Use(func(next geek_web.HandleFunc) geek_web.HandleFunc {
		return func(ctx *geek_web.Context) {
			ctime := time.Now()
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
		v1.GET("/user", func(ctx *geek_web.Context) {
			s := []int{1, 2, 3}
			fmt.Println(s[100]) // 这里直接报错：下标越界
			ctx.JSON(http.StatusOK, geek_web.H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"method": ctx.Method,
			})
		})
		v1.GET("/user/login", func(ctx *geek_web.Context) {
			ctx.JSON(http.StatusOK, geek_web.H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"method": ctx.Method,
			})
		})
		v1.GET("/assets/*filepath", func(ctx *geek_web.Context) {
			filePath, _ := ctx.Param("filepath")
			ctx.JSON(http.StatusOK, geek_web.H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"info":   "你是想访问我的这个文件吗？【" + filePath + "]",
				"method": ctx.Method,
			})
		})
		v1.GET("/user/:id/:action", func(ctx *geek_web.Context) {
			id, _ := ctx.Param("id")
			action, _ := ctx.Param("action")
			ctx.JSON(http.StatusOK, geek_web.H{
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
		v2.POST("/user", func(ctx *geek_web.Context) {
			ctx.JSON(http.StatusOK, geek_web.H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"method": ctx.Method,
			})
		})
		v2.POST("/user/login", func(ctx *geek_web.Context) {
			ctx.JSON(http.StatusOK, geek_web.H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"method": ctx.Method,
			})
		})
		v2.POST("/assets/*filepath", func(ctx *geek_web.Context) {
			filePath, _ := ctx.Param("filepath")
			ctx.JSON(http.StatusOK, geek_web.H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"info":   "你是想访问我的这个文件吗？【" + filePath + "]",
				"method": ctx.Method,
			})
		})
		v2.POST("/user/:id/:action", func(ctx *geek_web.Context) {
			id, _ := ctx.Param("id")
			action, _ := ctx.Param("action")
			ctx.JSON(http.StatusOK, geek_web.H{
				"code":   200,
				"msg":    "请求成功" + ctx.Pattern,
				"id":     id,
				"action": action,
				"method": ctx.Method,
			})
		})
	}
	v3 := s.Group("/v3")
	err := tplEngine.ParseGlob("../template/*.gohtml") // 用gohtml作为模板文件的后缀名，Goland会识别到这是一个模板文件
	if err != nil {
		panic("Web: 解析文件失败")
	}
	{
		v3.GET("/login", func(ctx *geek_web.Context) {
			data := struct {
				Username string
				Password string
			}{
				Username: "Neo",
				Password: "Neo123",
			}
			ctx.HTML(http.StatusOK, "login.gohtml", data)
		})
	}
	_ = s.Start(":8080")
}
