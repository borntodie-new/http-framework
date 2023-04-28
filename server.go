package geek_web

import (
	"fmt"
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
	router       *router // 路由树
	*RouterGroup         // 路由分组
}

// 这条语句没有任何实际作用，只是为了在语法层面上能够保证HTTPServer结构体实现了Server接口
var _ Server = &HTTPServer{}

// ServeHTTP  向前对接客户端请求，向后对接Web框架
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 构建上下文
	ctx := newContext(w, r)
	// 2. 匹配路由
	n, params, ok := s.findRouter(r.Method, r.URL.Path)
	if !ok || n.handler == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 NOT FOUND"))
		return
	}
	// 保存请求地址上的参数到上下文中
	ctx.params = params
	// 3. 执行命中路由的视图函数
	n.handler(ctx)
	// 4. 统一返回响应
	_ = ctx.resp()
}

func (s *HTTPServer) Start(addr string) error {
	// 直接使用内置方法启动一个服务，将HTTPServer作为IO多路复用器
	return http.ListenAndServe(addr, s)
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
func NewHTTPServer() *HTTPServer {
	router := newRouter()
	group := newRouterGroup()
	engine := &HTTPServer{
		router:      router,
		RouterGroup: group,
	}
	group.engine = engine
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
*/

type RouterGroup struct {
	prefix string       // 路由分组前缀
	parent *RouterGroup // 父路由组
	engine *HTTPServer  // server实例对象, 这样写有点不太优雅，因为这里应该是一个接口的，这样直接写成HTTPServer耦合性太高
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
	g.engine.router.addRouter(method, fmt.Sprintf("%s%s", g.prefix, pattern), handleFunc)
}

// findRouter 匹配路由
// 会有这个方法纯属是为了设计完整完整性，因为前面我们对于路由注册是完全在RouterGroup中完成的
// 由于完整性，我们也在RouterGroup中定义一个findRouter方法
func (g *RouterGroup) findRouter(method string, pattern string) (*node, map[string]string, bool) {
	return g.engine.router.findRouter(method, pattern)
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
	prefix = fmt.Sprintf("/%s", strings.Trim(prefix, "/"))
	newGroup := &RouterGroup{
		prefix: prefix,
		parent: g,
		engine: g.engine,
	}
	return newGroup
}

func newRouterGroup() *RouterGroup {
	return &RouterGroup{}
}
