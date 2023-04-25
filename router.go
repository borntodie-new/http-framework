package geek_web

import "strings"

// router 路由树的结构
// 1. 提供注册的功能
// 2. 提供匹配的功能
type router struct {
	// trees "路由树"，应该叫路由森林，因为每一个请求方法都会对应一颗树
	// 具体结构：{GET: tree, POST: tree, ...}
	trees map[string]*node
}

// newRouter 构造方法
func newRouter() *router {
	return &router{trees: map[string]*node{}}
}

// addRouter 注册路由
func (r *router) addRouter(method string, pattern string, handleFunc HandleFunc) {
	// 校验 pattern的相关信息
	// 1. 不能为空
	if pattern == "" {
		panic("Web: 路由不能是空字符串")
	}
	// 2. 不能以 / 结尾
	if pattern != "/" && strings.HasSuffix(pattern, "/") {
		panic("Web: 路由不能以 / 结尾")
	}
	// 3. 必须以 / 开头
	if !strings.HasPrefix(pattern, "/") {
		panic("Web: 路由必须以 / 开头")
	}

	root, ok := r.trees[method]
	if !ok {
		// 路由树不存在，直接创建并赋值
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	// 特殊处理跟路由
	if pattern == "/" {
		if root.handler != nil {
			// 为什么会路由冲突？
			// 正常来讲，当一个节点的handler有数据，就表示它是叶子节点，也表示他之前被创建过
			panic("Web: 路由冲突")
		}
		root.handler = handleFunc
		return
	}

	// 切割 pattern
	// pattern = /user/login
	// parts = ["", "user", "login"]
	// 第一个空串我们并不想要
	parts := strings.Split(pattern[1:], "/")
	// 一级一级添加节点
	for _, part := range parts {
		if part == "" {
			panic("Web: 不能注册连续 / 的路由")
		}
		root = root.childOrCreate(part)
	}
	if root.handler != nil {
		panic("Web: 路有冲突")
	}
	root.handler = handleFunc
}

// findRouter 匹配路由
func (r *router) findRouter(method string, pattern string) (*node, bool) {
	root, ok := r.trees[method]
	if !ok {
		// 不存在根路由树
		return nil, false
	}
	// 特殊处理根路由
	if pattern == "/" {
		return root, true
	}
	// 切割pattern
	// 还是需要将前面的 / 切割出去
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" {
			// 表示pattern是 /asud//asd/asd这种连续出现多个 / 情况
			// 其实这里可以不判断的，出现这样的情况，会切出 ""
			// 那 "" 肯定在路由树中找不到
			// 不过这里直接做判断，就省得在做无效功
			return nil, false
		}
		root, ok = root.childOf(part)
		if !ok {
			return nil, false
		}
	}
	// 这里我们也不能直接返回，还需要在进一步判断 当前找到的node节点的handler是否非nil，非nil才算成功
	return root, root.handler != nil
}

// node 树上节点的结构
// 匹配顺序
// 1. 静态匹配
// 2. 通配符匹配
type node struct {
	// path 单块的路径
	// /user/login => [user, login]
	// path = user
	path string

	// children 当前节点下所有的子节点
	children map[string]*node

	// handler 命中路由需要执行的逻辑
	// 只有叶子节点才会有这个属性
	handler HandleFunc

	// 通配符 * 表达的节点，任意匹配
	startChild *node
}

// childOf 用于匹配节点
// 查找节点，判断当前的节点的子节点中有没有path节点
func (n *node) childOf(part string) (*node, bool) {
	// 因为这里是查找，所以不存在当前节点的children属性是nil的情况
	// 只有一种情况会是这样，就是叶子节点
	child, ok := n.children[part]
	return child, ok
}

// childOrCreate 用于注册路由使用
// 查找节点，判断当前节点的子节点中是否存在path节点，已存在返回path节点，不存在就创建节点并添加到子节点中
func (n *node) childOrCreate(part string) *node {
	// 判断当前节点的子节点属性是否为nil
	// 为nil就创建
	if n.children == nil {
		n.children = map[string]*node{}
	}
	child, ok := n.children[part]
	if !ok {
		// part节点不存在，直接创建并添加
		child = &node{
			path: part,
		}
		n.children[part] = child
	}
	return child
}
