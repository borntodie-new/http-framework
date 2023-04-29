package accesslog

import (
	"encoding/json"
	geek_web "github.com/borntodie-new/geek-web"
	"log"
)

// 可观测性之日志记录

/*
在可观测性方案中主要考虑三个维度
	1. 日志记录
	2. 性能指标
	3. 链路追踪
对于这三个维度，我们最好是使用AOP来解决。好处如下
	1. 不会侵入我们的框架核心
	2. 不会侵入用户的业务核心
*/

// 实现日志记录功能
// 我们使用Builder设计模式，可扩展性高

// MiddlewareAccessLog 定义一个日志结构体，方便扩展
type MiddlewareAccessLog struct {
	// logFunc 函数类型，由于不清楚用户想要怎么处理这些日志信息
	// 可能是想写入文件
	// 可能是想输出控制台
	// ...
	// 我们这里暴露出去一个属性，用户自定义这个方法，具体是想往哪存都行
	logFunc func(information string)
}

func NewBuilder(logFunc func(information string)) *MiddlewareAccessLog {
	// 模式是输出到控制台
	if logFunc == nil {
		logFunc = func(information string) {
			log.Printf("%s\n", information)
		}
	}
	return &MiddlewareAccessLog{logFunc: logFunc}
}

// Builder Builder出一个中间件函数
func (m *MiddlewareAccessLog) Builder() geek_web.Middleware {
	return func(next geek_web.HandleFunc) geek_web.HandleFunc {
		return func(ctx *geek_web.Context) {
			defer func() {
				// 记录我们想要保存当前请求想要保存的信息
				// 如果我们想要记录命中的路由是什么，那从哪里拿呢？
				// 在这个作用域中，我们能和外界取得联系的只有Context上下文，所以我们只能通过Context获取到信息
				// 具体操作就是在Context中新增一个属性，在这里取就好
				l := accessLog{
					Host:    ctx.Request.Host,
					Method:  ctx.Method,
					Pattern: ctx.Pattern,
				}
				data, _ := json.Marshal(l)
				m.logFunc(string(data))
			}()
			// 注意：这里用了defer，对于defer有两个需要注意的点。这两个点只是针对于当前场景
			// 1. defer的执行时间是在下面的next方法执行完之后再执行的。也就是说执行完用户业务逻辑之后再执行
			// 2. defer可以防止next的执行过程中出现panic也能记录到信息
			next(ctx)
		}
	}
}

// accessLog 日志抽象结构体，可自定义
type accessLog struct {
	Host    string `json:"host"`    // 请求的主机地址
	Method  string `json:"method"`  // 请求的方法
	Pattern string `json:"pattern"` // 请求的路径
}

// Builder 下面这个表示这个中间件啥都不做
//func (m *MiddlewareAccessLog) Builder() geek_web.Middleware {
//	return func(next geek_web.HandleFunc) geek_web.HandleFunc {
//		return func(ctx *geek_web.Context) {
//			next(ctx)
//		}
//	}
//}
