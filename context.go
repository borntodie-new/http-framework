package geek_web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
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

	// 模板引擎对象
	t TemplateEngine

	// mu 加上读写锁，保护Keys信息
	mu sync.RWMutex
	// Keys 是一个键值对，实现中间件之间通信
	Keys map[string]any
	// 上面两个属性，我们是在实现session的时候，思考session怎么存储能够在其他地方能够直接获取到而来的灵感。
	// 其实对于Gin框架也是有这样的设计的。我们这也就刚好和Gin不谋而合。
	// 这个功能其实是非常常见的：就是怎么实现中间件的通信
	// 个人认为没有必要加上一个锁，这个锁的目的是为了保护Keys
	// Keys的作用是保存某些数据供其他中间件更方便使用
	// 中间件的执行是线性的，就是执行一个中间件，才能进入到下一个中间件
	// 应该是不存在并发的对Keys进行读写操作
	// 但是呢，人Gin使用了，那咱们也就加上一下吧，难度不是很大
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
		Method:   r.Method,
		Pattern:  r.URL.Path,
		header:   map[string]string{},
		status:   http.StatusOK, // 默认是200，因为状态码不能设置为0，但是int类型的零值是0
		data:     []byte(""),    // 响应体默认设置为空字符串吧，好像说响应体也不能为零值，对于这个还有点特殊，因为我们定义data类型是any类型的，也可以直接将Context中的data改成[]byte类型
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

// DelHeader 删除响应头数据
func (c *Context) DelHeader(key string) {
	delete(c.header, key)
}

// SetData 设置需要响应回去的数据
func (c *Context) SetData(data any) {
	c.data = data
}

// JSON 响应JSON格式数据
func (c *Context) JSON(code int, data any) {
	// 1. 设置状态码
	c.SetStatusCode(code)
	// 2. 设置响应头
	c.SetHeader("Content-Type", "application/json")
	// 3. 序列化
	bytes, err := json.Marshal(data)
	if err != nil {
		c.DelHeader("Context-Type")
		// 解析出错，直接panic，反正后面有recovery兜底
		panic(err)
	}
	c.SetData(bytes)
}

// HTML TODO 响应HTML格式数据回去
func (c *Context) HTML(code int, templateName string, data any) {
	// 1. 设置状态码
	c.SetStatusCode(code)
	// 2. 设置响应头
	c.SetHeader("Content-Type", "text/html")
	// 3. 设置响应数据
	html, err := c.t.Render(c, templateName, data)
	if err != nil {
		c.SetStatusCode(http.StatusInternalServerError)
		c.DelHeader("Content-Type")
		html = []byte("Server Internal Error, Please Try Again Later!")
	}
	c.SetData(html)
}

// String 响应纯文本的数据回去
func (c *Context) String(code int, data []byte) {
	// 1. 设置状态码
	c.SetStatusCode(code)
	// 2. 设置响应头
	c.SetHeader("Content-Type", "text/plain")
	// 3. 设置响应数据
	c.SetData(data)
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

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}

	c.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}

// MustGet 特别注意下个方法，如果没有获取到信息，会直接panic的
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *Context) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *Context) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Context) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetUint returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint(key string) (ui uint) {
	if val, ok := c.Get(key); ok && val != nil {
		ui, _ = val.(uint)
	}
	return
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint64(key string) (ui64 uint64) {
	if val, ok := c.Get(key); ok && val != nil {
		ui64, _ = val.(uint64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime returns the value associated with the key as time.
func (c *Context) GetTime(key string) (t time.Time) {
	if val, ok := c.Get(key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func (c *Context) GetDuration(key string) (d time.Duration) {
	if val, ok := c.Get(key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (c *Context) GetStringSlice(key string) (ss []string) {
	if val, ok := c.Get(key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (c *Context) GetStringMap(key string) (sm map[string]any) {
	if val, ok := c.Get(key); ok && val != nil {
		sm, _ = val.(map[string]any)
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (c *Context) GetStringMapString(key string) (sms map[string]string) {
	if val, ok := c.Get(key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
	if val, ok := c.Get(key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}

// 下述方法这只是临时性的，我们后面会通过AOP的形式完成自动设置状态码、响应头和响应数据到Response中

//// setHeader 统一设置响应头
//func (c *Context) setHeader() {
//	for k, v := range c.header {
//		c.Response.Header().Set(k, v)
//	}
//}
//
//// setStatusCode 统一设置状态码
//func (c *Context) setStatusCode() {
//	c.Response.WriteHeader(c.status)
//}
//
//// setData 统一设置响应数据
//func (c *Context) setData() error {
//	_, err := c.Response.Write((c.data).([]byte))
//	return err
//}

// Resp 统一返回
// 目前这个好像得暴露出去，不然在同意刷新数据到响应体的中间件中就不能调用了
// 其实好的就是小写，使其成为一个内部方法
//func (c *Context) Resp() error {
//	c.setHeader()
//	c.setStatusCode()
//	return c.setData()
//}

// resp 因为flashData内部中间件和Context上下文在同一个包中，所有不再需要将下面的方法改成大写暴露出去
//func (c *Context) resp() error {
//	c.setHeader()
//	c.setStatusCode()
//	return c.setData()
//}

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
