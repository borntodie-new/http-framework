package cookie

import (
	"github.com/borntodie-new/geek-web/session"
	"net/http"
)

// 实现Propagator接口
// 主要功能是向前对接请求，向后对接视图函数

var _ session.Propagator = &Propagator{}

// PropagatorOpt Propagator可选项配置，其实是对cookie的可选项配置
type PropagatorOpt func(propagator *Propagator)

type Propagator struct {
	// cookieName 设置在响应体中的cookie的键的名字
	cookieName string
	// cookieOpt 个人感觉这个其实没有必要，因为这个结构体已经和cookie高度耦合了，再抽离出一个这样的逻辑没啥必要
	// 可能会有朋友认为，我们作为框架的设计者，不太清楚developer到底想将session存储在Response的那个位置？
	// 这个其实是多虑了，因为咱们这个文件就是基于cookie将session发送给客户端，从客户端请求的cookie中拿解析session信息
	// cookieOpt func(cookie *http.Cookie)

	// 上面对于cookieOpt的想法是错的，我完全没理解对这个功能的作用，我一直是以为cookieOpt是对cookie存储位置进行配置的
	// 其实不然，这是对具体cookie的配置的，比如：cookie的过期时间，cookie的生效域名、cookie的安全配置等等信息
	// 对于上述对cookie的配置都是一些可选项，我们框架最基本最基本要做的就是配置cookie的name和value
	cookieOpt func(cookie *http.Cookie)
}

// WithCookieOpt 框架层面提供一个配置方法
func WithCookieOpt(cookieOpt func(cookie *http.Cookie)) PropagatorOpt {
	return func(propagator *Propagator) {
		propagator.cookieOpt = cookieOpt
	}
}

// NewPropagator 初始化一个Propagator
// 这里接受一个可扩展的参数其实有点过度设计了
// 因为目前只是需要对cookie进行配置
// 但其实也还能接受
func NewPropagator(cookName string, opts ...PropagatorOpt) *Propagator {
	p := &Propagator{
		cookieName: cookName,
		cookieOpt:  func(cookie *http.Cookie) {},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Inject 将id保存到response的cookie中
// 其实将session的唯一标识叫做id在这里其实不太好，有点误导人
// 对接客户端
func (p *Propagator) Inject(id string, response http.ResponseWriter) error {
	// 这里只是设置cookie最基本最基本的信息
	cookie := &http.Cookie{
		Name:  p.cookieName,
		Value: id,
	}
	// 对cookie进行用户配置用户选项
	p.cookieOpt(cookie)
	// 将cookie设置到response中
	http.SetCookie(response, cookie)
	return nil
}

// Extract 从request的cookie中解析出session
// 对接视图，也就是对接developer
func (p *Propagator) Extract(request *http.Request) (string, error) {
	cookie, err := request.Cookie(p.cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Remove 将cookie从response中删除，其实就是使之前的cookie失效
func (p *Propagator) Remove(response http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:   p.cookieName,
		MaxAge: -1, // 关键使这里，并且
	}
	http.SetCookie(response, cookie)
	return nil
}
