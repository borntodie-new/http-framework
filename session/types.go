package session

import (
	"context"
	"net/http"
)

// Propagator 决定session的保存和提取
type Propagator interface {
	// Inject 往哪注入session
	// id 表示当前需要存储的session唯一标识
	// writer 响应对象
	Inject(id string, response http.ResponseWriter) error

	// Extract 从哪提取出session信息
	Extract(request *http.Request) (string, error)

	// Remove 从Response中移除session
	Remove(response http.ResponseWriter) error
}

// Session 具体决定session的存储位置
// 表现为对session的存取
// 可能有些朋友可能对ID、key、value之间是什么关系？
// {
// 	ID1: {
// 		key: value
// 	},
// 	ID2: {
// 		key: value
// 	},
// 	ID3: {
// 		key: value
// 	},
// }
// 整个map就是一个store，每个ID就是一个session，每个session中其实可以存储多个数据信息。这个其实是根据具体的实现来定的。
type Session interface {
	// Get 获取Store中key所对应的value
	Get(ctx context.Context, key string) (any, error)
	// Set 设置Store中设置一个键值对，键是key，值是value
	Set(ctx context.Context, key string, value any) error
	// ID 获取当前session的唯一标识
	ID() string
}

// Store 管理所有的session
// 向前对接developer，向后对接Session
type Store interface {
	// Generate 创建一个session对象
	Generate(ctx context.Context, id string) (Session, error)
	// Refresh 刷新session对象
	Refresh(ctx context.Context, id string) error
	// Remove 删除session对象
	Remove(ctx context.Context, id string) error
	// Retrieve 获取session对象
	Retrieve(ctx context.Context, id string) (Session, error)
}
