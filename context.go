package geek_web

import "net/http"

// Context 请求响应的上下文，请求过来到响应回去的全过程
// 那这个需要抽象成一个接口吗？
// 其实没必要，我们需要抽象成接口的一般都是为了解决扩展性问题，而对于上下文一般不会有很大的改动
type Context struct {
	// Request 请求。没有自行封装的必要，因为一个请求过来了，一般里面的数据都不会变化
	Request *http.Request
	// Response 响应。建议自行封装，为了扩展性
	Response http.ResponseWriter
}
