package geek_web

import "net/http"

// HandleFunc 视图函数的唯一签名
type HandleFunc func(ctx *Context)

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
	*router
}

// 这条语句没有任何实际作用，只是为了在语法层面上能够保证HTTPServer结构体实现了Server接口
var _ Server = &HTTPServer{}

// ServeHTTP  向前对接客户端请求，向后对接Web框架
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
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

func (s *HTTPServer) GET(pattern string, handleFunc HandleFunc) {
	s.addRouter(http.MethodGet, pattern, handleFunc)
}

// NewHTTPServer 构造方法
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
	}
}
