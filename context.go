package geek_web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Context 请求响应的上下文，请求过来到响应回去的全过程
// 那这个需要抽象成一个接口吗？
// 其实没必要，我们需要抽象成接口的一般都是为了解决扩展性问题，而对于上下文一般不会有很大的改动
type Context struct {
	// Request 请求。没有自行封装的必要，因为一个请求过来了，一般里面的数据都不会变化
	Request *http.Request
	// Response 响应。建议自行封装，为了扩展性
	Response http.ResponseWriter
	// 路由参数或通配符参数
	params map[string]string

	// 缓存请求地址
	cacheQuery url.Values
	// 缓存请求体
	cacheBody io.ReadCloser
	// request: 常用的信息
	Method  string // 请求方法
	Pattern string // 请求地址

	// response: 常用的信息
	status int               // 状态码
	data   any               // 需要响应回去的数据：任意数据
	header map[string]string // 响应头数据
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
		Method:   r.Method,
		Pattern:  r.URL.Path,
		header:   map[string]string{},
	}
}

// Param 获取请求地址上的参数
// /user/:id
// /user/123
// Param(id) => 123
func (c *Context) Param(key string) (string, error) {
	value, ok := c.params[key]
	if !ok {
		return "", errors.New(fmt.Sprintf("Web: %s 不存在", key))
	}
	return value, nil
}

// SetStatusCode 设置响应状态码
func (c *Context) SetStatusCode(code int) {
	c.status = code
}

// SetHeader 设置响应头
func (c *Context) SetHeader(key, value string) {
	c.header[key] = value
}

// SetData 设置需要响应回去的数据
func (c *Context) SetData(data any) {
	c.data = data
}

// JSON 响应JSON格式数据
func (c *Context) JSON(code int, data any) error {
	// 1. 设置状态码
	c.SetStatusCode(code)
	// 2. 设置响应头
	c.SetHeader("Content-Type", "application/json")
	// 3. 序列化
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.SetData(bytes)
	return c.resp()
}

// HTML TODO 响应HTML格式数据回去
func (c *Context) HTML(code int, data any) error {
	// 1. 设置状态码
	c.SetStatusCode(code)
	// 2. 设置响应头
	c.SetHeader("Content-Type", "text/html")
	// 3. 设置响应数据
	// TODO
	return c.resp()
}

// String 响应纯文本的数据回去
func (c *Context) String(code int, data []byte) error {
	// 1. 设置状态码
	c.SetStatusCode(code)
	// 2. 设置响应头
	c.SetHeader("Content-Type", "text/plain")
	// 3. 设置响应数据
	c.SetData(data)
	return c.resp()
}

// Query 获取查询参数
func (c *Context) Query(key string) (string, error) {
	if c.cacheQuery == nil {
		c.cacheQuery = c.Request.URL.Query()
	}
	value, ok := c.cacheQuery[key]
	if !ok {
		return "", errors.New(fmt.Sprintf("Web: %s不存在", key))
	}
	return value[0], nil
}

// Form 解析请求体数据
// 注意：我们这只是将请求的body缓存起来了，方便需要多次从请求体中获取数据
// 一般我们获取请求体数据，还都是从原来request的body中获取
func (c *Context) Form(key string) (string, error) {
	if c.cacheBody == nil {
		c.cacheBody = c.Request.Body
	}
	// 必须先使用ParseForm方法
	if err := c.Request.ParseForm(); err != nil {
		return "", err
	}
	// 从请求体中解析数据
	return c.Request.FormValue(key), nil
}

// SetCookie 这种方法其实没必要封装，因为http包已经内置了一个非常简单的方法
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

// 下述方法这只是临时性的，我们后面会通过AOP的形式完成自动设置状态码、响应头和响应数据到Response中

// setHeader 统一设置响应头
func (c *Context) setHeader() {
	for k, v := range c.header {
		c.Response.Header().Set(k, v)
	}
}

// setStatusCode 统一设置状态码
func (c *Context) setStatusCode() {
	c.Response.WriteHeader(c.status)
}

// setData 统一设置响应数据
func (c *Context) setData() error {
	_, err := c.Response.Write((c.data).([]byte))
	return err
}

// resp 统一返回
func (c *Context) resp() error {
	c.setHeader()
	c.setStatusCode()
	return c.setData()
}

// - 思考：我们在Param方法、Query方法、Form方法，我们需不需要帮用户解析好
// 例如：
//	/user/:id
//
//	id, _ := ctx.Param("id")
//	返回回来的id是一个字符串
//	我的意思是需不需要将这个字符串转成整型
//	其实不太建议，如果还是想做，可以考虑下面这种方式
//
//func (c *Context) Query(key string) StringValue {
//	if c.cacheQuery == nil {
//		c.cacheQuery = c.Request.URL.Query()
//	}
//	value, ok := c.cacheQuery[key]
//	if !ok {
//		return StringValue{err: errors.New(fmt.Sprintf("Web: %s不存在", key))}
//	}
//	return StringValue{value: value[0]}
//}
//
//type StringValue struct {
//	value string
//	err   error
//}
//
//func (s StringValue) String() (string, error) {
//	return s.value, s.err
//}
//func (s StringValue) ToInt64() (int64, error) {
//	if s.err != nil {
//		return 0, s.err
//	}
//	return strconv.ParseInt(s.value, 10, 64)
//}
