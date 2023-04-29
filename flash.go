package geek_web

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
func (m *MiddlewareFlashData) Builder() Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			defer func() {
				// 统一刷新数据到response中
				// 1. 设置响应头
				for k, v := range ctx.header {
					ctx.Response.Header().Set(k, v)
				}
				// 2. 设置状态码
				ctx.Response.WriteHeader(ctx.status)
				// 3. 设置响应体
				_, err := ctx.Response.Write((ctx.data).([]byte))
				// 如果刷新数据到响应体中出现错误，直接panic
				// 后面会有一个recovery hook住panic错误的
				if err != nil {
					panic(err)
				}
				//_ = ctx.resp()
			}()
			next(ctx)
		}
	}
}

// FlashDataBuilder 初始化结构体
func FlashDataBuilder() *MiddlewareFlashData {
	return &MiddlewareFlashData{}
}
