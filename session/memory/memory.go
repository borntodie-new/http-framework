package memory

import (
	"context"
	"fmt"
	"github.com/borntodie-new/geek-web/session"

	"time"

	"github.com/patrickmn/go-cache"
)

// Store和Session两个对象其实是强耦合的，所以一般来说都是配套使用，配套定义的

// 代码层面约束类型
var _ session.Session = &Session{}
var _ session.Store = &Store{}

// Session 对Session结构的定义。决定session中到底存啥
// 看到下面这个Session结构体，不知道大家有没有疑问
// 就是为什么data要定义成一个map，不能是直接一个简单的数据类型吗？比如说string类型
// 这其实是接口那边就定义好了，大家可以自定看下`types`文件中的Session接口定义
type Session struct {
	// id 当前session的唯一标识ID
	id string
	// data 当前session存储的具体数据
	data map[string]string
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	value, ok := s.data[key]
	if !ok {
		return "", fmt.Errorf("web: NOT FOUND this key %s", key)
	}
	return value, nil
}

func (s *Session) Set(ctx context.Context, key string, value any) error {
	s.data[key] = value.(string)
	return nil
}

func (s *Session) ID() string {
	return s.id
}

// Store 对Session进行集中管理
type Store struct {
	// c session的存储位置，利用缓存来帮我们管理session
	c *cache.Cache
	// expiration session过期时间
	expiration time.Duration
}

func NewStore(expiration time.Duration) *Store {
	return &Store{
		// 第一个参数是存储的数据有效期
		// 第二个参数是过期的数据隔多长时间清理 隔 1s 清理
		c:          cache.New(expiration, time.Second),
		expiration: expiration,
	}
}

// Generate 生成session
func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	// 创建session对象
	sess := &Session{
		id:   id,
		data: make(map[string]string),
	}
	// 这里的过期时间是session的有效时间，不是到期时间
	// 比如：有效期是30分钟
	// Set方法内部已经实现了从当前时间加上expiration
	s.c.Set(id, sess, s.expiration)
	return sess, nil
}

// Refresh 刷新session
// 刷新session有两种方法
// 1. 重新设置原来的那个session，不过修改了过期时间
// 2. 重新生成一个session，设置到缓存中
// 我们这里用的是方式1：
// 原因也是在`Store`接口中定义好的
func (s *Store) Refresh(ctx context.Context, id string) error {
	sess, err := s.Retrieve(ctx, id)
	if err != nil {
		return fmt.Errorf("refresh current session is error")
	}
	s.c.Set(sess.ID(), sess, s.expiration)
	return nil
}

// Remove 删除session
func (s *Store) Remove(ctx context.Context, id string) error {
	s.c.Delete(id)
	return nil
}

// Retrieve 获取session
func (s *Store) Retrieve(ctx context.Context, id string) (session.Session, error) {
	sess, ok := s.c.Get(id)
	if !ok {
		return nil, fmt.Errorf("current session is not exist")
	}
	return sess.(*Session), nil
}
