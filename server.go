package geek_web

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// HandleFunc 视图函数的唯一签名
type HandleFunc func(ctx *Context)

// H 提供一个map结构，方便用户操作
type H map[string]any

// Server 接口
// 为什么要这么设计，我们直接一个结构体实现http.Handler接口不可以吗
// 是可以的，但是为了兼容以后的HTTPS协议做准备
type Server interface {
	// Handler 组装http.Handler接口，确保这个接口能够实现Server功能，也就是能够充当一个IO多路复用器
	http.Handler

	// Start 作为Server启动的入口
	Start(addr string) error

	// AddRouter 注册路由的唯一方法
	// method 请求方法
	// path URL 路径，必须以 / 开头
	// handlerFunc 视图函数
	// 这是内部核心的API，没必要暴露出去，所以改成小写
	addRouter(method string, path string, handleFunc HandleFunc)
}

// HTTPServer 实现一个HTTP协议的Server接口
type HTTPServer struct {
	router       *router        // 路由树
	*RouterGroup                // 路由分组
	groups       []*RouterGroup // 保存程序中产生的所有路由组实例
	// templateEngine 这里只是为了一个过渡，最终还是或将这个落到Context上下文中
	// 我们思考一下，这个模板渲染的功能是所有的用户都需要的吗？或者说，至少大部分用户都需要用到？
	// 其实不是的，这个功能对很多用户来说并不需要，所以我们这里可以做一个优化处理，对于有需求的用户，需要额外再做一些配置，对HTTPServer对象
	templateEngine TemplateEngine
}

// ServerOption 抽象一个可配置的类型
type ServerOption func(server *HTTPServer)

// ServerWithTemplateEngine 给HTTPServer配置配置上模板引擎对象
// 需要向外暴露，如果有需求的用户直接使用这个可配置项方法进行配置
// 这种方法有点像装饰器😀😀
func ServerWithTemplateEngine(t TemplateEngine) ServerOption {
	return func(server *HTTPServer) {
		server.templateEngine = t
	}
}

// 这条语句没有任何实际作用，只是为了在语法层面上能够保证HTTPServer结构体实现了Server接口
var _ Server = &HTTPServer{}

//func (s *HTTPServer) server(ctx *Context) {
//	// 2. 匹配路由
//	n, params, ok := s.findRouter(ctx.Method, ctx.Pattern)
//	if !ok || n.handler == nil {
//		// 简陋了一点，直接返回错误信息到前端用户
//		ctx.Response.WriteHeader(http.StatusNotFound)
//		_, _ = ctx.Response.Write([]byte("404 NOT FOUND"))
//		return
//	}
//	// 保存请求地址上的参数到上下文中
//	ctx.params = params
//	// 3. 执行命中路由的视图函数
//	n.handler(ctx)
//	// 4. 统一返回响应
//	_ = ctx.resp()
//}

// ServeHTTP  向前对接客户端请求，向后对接Web框架
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 构建上下文
	ctx := newContext(w, r)
	// 将HTTPServer中的TemplateEngine对象转给Context上下文对象
	ctx.t = s.templateEngine
	log.Printf("REQUEST COMING %4s - %s", ctx.Method, ctx.Pattern)
	// 2. 匹配路由
	n, params, ok := s.findRouter(ctx.Method, ctx.Pattern)
	if !ok || n.handler == nil {
		// 下面的逻辑目前是直接写数据到响应体中，并且直接返回到客户端
		// 不太好，因为这种方式没有执行框架内部中间件和用户中间件
		//w.WriteHeader(http.StatusNotFound)
		//_, _ = w.Write([]byte("404 NOT FOUND"))
		//return

		// 优化: 主要思路就是将n【匹配到的路由节点】篡改掉，具体是改handler方法
		n = &node{handler: func(ctx *Context) {
			ctx.SetStatusCode(http.StatusNotFound)
			ctx.SetData([]byte("404 NOT FOUND"))
			return
		}}
	}
	// 保存请求地址上的参数到上下文中
	ctx.params = params
	// 匹配路由组 ---> 获取中间件
	middlewares := s.filterGroup(ctx.Pattern)
	if middlewares == nil {
		middlewares = make([]Middleware, 0)
	}
	// 重点：将命中的视图函数添加到中间件列表中，统一执行
	// 必须倒序组装，只有这样，最后出来的handler才会是第一个注册的中间件
	// handler其实一直在变
	// 具体原理参考文章：https://juejin.cn/post/7227139379105038392
	handler := n.handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler) // 第一次执行的时候，handler其实还是用户的业务视图
	}

	// 没执行下面的方法之前，handler是用户注册的第一个中间件函数
	// 执行下面的方法之后，handler就是框架内部注册的第一个中间件函数

	// 这里统一注册框架内部的中间件
	handler = s.registerMiddlewares(handler) // registerMiddlewares方法必须是将用户的中间件注册完之后才能注册框架内部的中间件逻辑
	// 具体执行全部的中间件
	handler(ctx) // 此时handler已经变了

	//// 具体执行中间件方法
	//handler(ctx) // 此时handler是用户注册的第一个中间件

	// 3. 执行命中路由的视图函数
	// n.handler(ctx)
	// 4. 统一返回响应
	// _ = ctx.Resp()
}

func (s *HTTPServer) Start(addr string) error {
	// 直接使用内置方法启动一个服务，将HTTPServer作为IO多路复用器
	return http.ListenAndServe(addr, s)
}

func (s *HTTPServer) filterGroup(pattern string) []Middleware {
	for _, group := range s.groups {
		if strings.HasPrefix(pattern, group.prefix) {
			return group.middlewares
		}
	}
	return nil
}

// registerMiddlewares 注册框架内部的中间件
func (s *HTTPServer) registerMiddlewares(handler HandleFunc) HandleFunc {
	for _, middleware := range s.initInternalMiddlewares() {
		handler = middleware(handler)
	}
	return handler
}

// initInternalMiddlewares 初始化注册框架内部的中间件
// 这里的注册逻辑是倒序注册的，越是最外层的越加最晚注册到middlewares中
func (s *HTTPServer) initInternalMiddlewares() []Middleware {
	middlewares := make([]Middleware, 0)
	// 这种注册方式不太好，如果内部需要注册100个中间件，那我们需要手动写100次吗？
	// 目前暂时用这种方式吧。
	// 其实可以用IOC方案

	// 注册recovery中间件
	middlewares = append(middlewares, RecoveryBuilder(nil).Builder())
	// 注册刷新数据中间件
	middlewares = append(middlewares, FlashDataBuilder().Builder())
	return middlewares
}

// addRouter 作为注册路由的唯一通道
// 疑问1：路由存在哪里？
// 疑问2：路由以怎样的结构存储？
// 因为*router嵌到HTTPServer中了，当*router实现了addRouter方法，也就表示HTTPServer实现了addRouter方法，不过这样做耦合性高
//func (s *HTTPServer) addRouter(method string, path string, handleFunc HandleFunc) {
//	s.addRouter(method, path, handleFunc)
//}

//func (s *HTTPServer) GET(pattern string, handleFunc HandleFunc) {
//	s.addRouter(http.MethodGet, pattern, handleFunc)
//}
//
//func (s *HTTPServer) POST(pattern string, handleFunc HandleFunc) {
//	s.addRouter(http.MethodPost, pattern, handleFunc)
//}
//
//func (s *HTTPServer) DELETE(pattern string, handleFunc HandleFunc) {
//	s.addRouter(http.MethodDelete, pattern, handleFunc)
//}
//
//func (s *HTTPServer) PUT(pattern string, handleFunc HandleFunc) {
//	s.addRouter(http.MethodPut, pattern, handleFunc)
//}

// NewHTTPServer 构造方法
// server和RouterGroup是相互应用了
func NewHTTPServer(opts ...ServerOption) *HTTPServer {
	r := newRouter()
	group := newRouterGroup()
	engine := &HTTPServer{
		router:      r,
		RouterGroup: group,
		groups:      []*RouterGroup{},
	}
	group.engine = engine
	// 通过这个就能做成一个可配置的HTTPServer了
	for _, opt := range opts {
		opt(engine)
	}
	return engine
}

/*
- 思考：
	1. 路由分组怎么实现？
	2. 路由分组应该怎么用？
	3. 路由分组结构怎么设计？

初步思考结论
	1. 只能有一个路由树，也就是说只能有一个router实例
	2. 可以有多个路由分组
	3. 综上：一个server需要内嵌一个router路由树，包含一个RouterGroup路由分组
	4. RouterGroup需要内嵌一个Server
	5. 所有的衍生API都是在RouterGroup中完成

设计中间件
	我们的设计是，实现一个路由组上的中间件，因为每个组可以注册一类中间件。不同的分组间中间件相互隔离的
	正是因为这个路由分组中的中间件是相互隔离的，所以我们不能统一管理这些中间
	我们需要重新定义一个结构保存整个程序中产生的所有路由组，有了路由组，就表示所有的中间件就都有了
那问题来了，所有的路由组保存在哪里呢？
	所有的路由组只需要保存一份就好，所以server是最好的选择
*/

type RouterGroup struct {
	prefix      string       // 路由分组前缀
	parent      *RouterGroup // 父路由组
	engine      *HTTPServer  // server实例对象, 这样写有点不太优雅，因为这里应该是一个接口的，这样直接写成HTTPServer耦合性太高
	middlewares []Middleware // 全部的中间件。注意，这里的middlewares是保存着当前路由组这条线上所有的中间件
}

func (g *RouterGroup) GET(pattern string, handleFunc HandleFunc) {
	g.addRouter(http.MethodGet, pattern, handleFunc)
}

func (g *RouterGroup) POST(pattern string, handleFunc HandleFunc) {
	g.addRouter(http.MethodPost, pattern, handleFunc)
}

func (g *RouterGroup) DELETE(pattern string, handleFunc HandleFunc) {
	g.addRouter(http.MethodDelete, pattern, handleFunc)
}

func (g *RouterGroup) PUT(pattern string, handleFunc HandleFunc) {
	g.addRouter(http.MethodPut, pattern, handleFunc)
}

// addRouter 注册路由
// 唯一和路由树做交互的通道
func (g *RouterGroup) addRouter(method string, pattern string, handleFunc HandleFunc) {
	pattern = fmt.Sprintf("%s%s", g.prefix, pattern)
	g.engine.router.addRouter(method, pattern, handleFunc)
	log.Printf("REGISTER ROUTER %4s - %s", method, pattern)
}

// findRouter 匹配路由
// 会有这个方法纯属是为了设计完整完整性，因为前面我们对于路由注册是完全在RouterGroup中完成的
// 由于完整性，我们也在RouterGroup中定义一个findRouter方法
func (g *RouterGroup) findRouter(method string, pattern string) (*node, map[string]string, bool) {
	return g.engine.router.findRouter(method, pattern)
}

// Group 创建路由分组
// 1. 创建一个新的路由分组
// 2. 将新的路由分组添加到路由组中央（server的groups属性中）
func (g *RouterGroup) Group(prefix string) *RouterGroup {
	prefix = fmt.Sprintf("/%s", strings.Trim(prefix, "/"))
	newGroup := &RouterGroup{
		prefix:      prefix,
		parent:      g,
		engine:      g.engine,
		middlewares: g.middlewares,
	}
	g.engine.groups = append(g.engine.groups, newGroup)
	return newGroup
}

// Use 注册中间件
// 将中间件保存在路由组中
func (g *RouterGroup) Use(middlewares ...Middleware) {
	if g.middlewares == nil {
		g.middlewares = middlewares
		return
	}
	g.middlewares = append(g.middlewares, middlewares...)
}

func newRouterGroup() *RouterGroup {
	return &RouterGroup{}
}
