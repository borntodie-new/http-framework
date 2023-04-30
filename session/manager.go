package session

import (
	"github.com/borntodie-new/geek-web"
)

// manager完全是出于用户[developer]体验考虑的
// 主要的功能就是集中管理Store和Propagator

type Manager struct {
	Store
	Propagator
	SessKey string // 将session设置到Context的Keys中的key的名字
}

// CreateSession 创建session对象并将session设置到cookie中
// 这个创建过程其实有点像下面我描述的这个过程
// 具体的session像我的左手，id像我右手
// 这里是把我的右手给到你
// 但是我的左手还是自由的，还是可以动的，甚至是修改session的信息
// 只要下次你通过我的右手还是可以连接到我的左手的
// 不知道这样讲大家有没有理解
func (m *Manager) CreateSession(ctx *geek_web.Context, id string) (Session, error) {
	// 1. 创建一个session对象
	sess, err := m.Generate(ctx.Request.Context(), id)
	if err != nil {
		return nil, err
	}
	// 2. 将session设置到cookie中
	err = m.Inject(sess.ID(), ctx.Response)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// DeleteSession 删除session对象并将session冲cookie中删除
// 这里的过程就像，直接将左右手强制分离
// 让左右受没有任何联系
func (m *Manager) DeleteSession(ctx *geek_web.Context, id string) error {
	// 1. 先删除Store中的session
	err := m.Store.Remove(ctx.Request.Context(), id)
	if err != nil {
		return err
	}
	// 2. 删除Propagator中的session，也就是将Response中设置的session删除
	err = m.Propagator.Remove(ctx.Response)
	if err != nil {
		return err
	}
	return nil
}

// UpdateSession 刷新session对象并将session重新设置到cookie中
// 下面步骤有点麻烦了，其实完全可以实现这样一个刷新session机制: 刷新session就会重新生成一个新的session
// 但是由于接口已经设计好了，我们就不变了😂😂😂
func (m *Manager) UpdateSession(ctx *geek_web.Context, id string) (Session, error) {
	// 1. 先查询session唯一标识是id的session对象
	sess, err := m.RetrieveSession(ctx)
	if err != nil {
		return nil, err
	}
	// 2. 更新Store中的session对象
	err = m.Refresh(ctx.Request.Context(), sess.ID())
	if err != nil {
		return nil, err
	}
	// 3. 更新Propagator中的session数据
	err = m.Inject(id, ctx.Response)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// RetrieveSession 获取session对象
// 1. 先从上下文的Keys中拿
// 2. Keys中没有，从Propagator中拿
// 3. 拿到了就设置到上下文的Keys中
func (m *Manager) RetrieveSession(ctx *geek_web.Context) (Session, error) {
	value, exists := ctx.Get(m.SessKey)
	if exists {
		return value.(Session), nil
	}
	id, err := m.Extract(ctx.Request) // 这里返回的ID的存储在Store中的session的key
	if err != nil {
		return nil, err
	}
	sess, err := m.Retrieve(ctx.Request.Context(), id)
	if err != nil {
		return nil, err
	}
	// 将session设置到Context的Keys中
	ctx.Set(m.SessKey, sess)
	return sess, nil
}
