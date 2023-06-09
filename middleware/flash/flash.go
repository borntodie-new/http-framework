package flash

import "github.com/borntodie-new/geek-web"

// TODO 本来是想在middleware中写这个逻辑的，但是出现了循环引入的问题

// 会有这个需求的原因
// 1. 由于http.ResponseWriter的特殊机制，一旦向响应体中写入了数据，后续再想写入或者修改就无效了
// 2. 统一写入更加友好，不会每写一个返回都需要手动向响应体中写入，要让用户对这一步骤无感

// 思考1
// 这个的刷新的功能是什么时候执行
// 1. 应该是在"最外层"，至少是远离业务逻辑的
// 2. 请求来的时候不需要做任何事情，只有在请求走了之后需要刷新数据

// 思考2
// 像这类框架内部需要注册的中间件，怎么注册呢？
// 1. 在RouterGroup创建的时候注册
// 2. 在匹配路由组的时候在其中间件最前面添加上内部的中间件

// MiddlewareFlashData 统一刷新数据到Response中
type MiddlewareFlashData struct{}

// Builder 构建中间件函数
func (m *MiddlewareFlashData) Builder() geek_web.Middleware {
	return func(next geek_web.HandleFunc) geek_web.HandleFunc {
		return func(ctx *geek_web.Context) {
			defer func() {
				// 统一刷新数据到response中
				_ = ctx.Resp()
			}()
			next(ctx)
		}
	}
}

// Builder 初始化结构体
func Builder() *MiddlewareFlashData {
	return &MiddlewareFlashData{}
}
