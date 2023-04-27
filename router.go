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
			part: "/",
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
// 我们想一想：
// 1. 对于精确匹配：我们需要做的操作不多，只需要将匹配到的节点返回即可
// 2. 对于通配符匹配：我们不支持这种路由：/assets/*filepath/jason，也就是说，当一个路由中出现了*，就表示匹配到了第一个节点就返回
// 3. 对于参数匹配：我们需要支持这种路由：/user/:id/update，也就是说，当一个路由中出现了 : ，就表示还得按照精确匹配的逻辑，一直执行匹配下去
func (r *router) findRouter(method string, pattern string) (*node, map[string]string, bool) {
	// 保存参数路径参数
	params := make(map[string]string)
	root, ok := r.trees[method]
	if !ok {
		// 不存在根路由树
		return nil, params, false
	}
	// 特殊处理根路由
	if pattern == "/" {
		return root, params, true
	}
	// 切割pattern
	// 还是需要将前面的 / 切割出去
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		pOk := false
		if part == "" {
			// 表示pattern是 /asud//asd/asd这种连续出现多个 / 情况
			// 其实这里可以不判断的，出现这样的情况，会切出 ""
			// 那 "" 肯定在路由树中找不到
			// 不过这里直接做判断，就省得在做无效功
			return nil, params, false
		}
		root, pOk, ok = root.childOf(part)
		if !ok {
			return nil, params, false
		}
		if pOk { // 是否是参数匹配 :和*匹配
			if root.starChild != nil { // 匹配到 * 节点，贪婪匹配之后直接返回
				// 这里对于 * 匹配是贪婪匹配，就是说后面的所有路径都要，表示这里需要直接 return
				// 问题是如何做到贪婪匹配?
				// 注册的路由：/assets/*filepath
				// 请求的路由：/assets/css/neo.css
				// 现在用filepath作为key
				// 现在用css/neo.css作为value
				index := strings.Index(pattern, part)
				params[root.part[1:]] = pattern[index:]
				return root, params, true
			}
			// 参数路由 : ，匹配到还需要继续往下查找
			// 并且需要记录好参数
			params[root.part[1:]] = part
		}
	}
	// 这里我们也不能直接返回，还需要在进一步判断 当前找到的node节点的handler是否非nil，非nil才算成功
	return root, params, root.handler != nil
}

// node 树上节点的结构
// 匹配顺序
// 1. 静态匹配
// 2. 通配符匹配
type node struct {
	// part 单块的路径
	// /user/login => [user, login]
	// part = user
	part string

	// children 当前节点下所有的子节点
	children map[string]*node

	// handler 命中路由需要执行的逻辑
	// 只有叶子节点才会有这个属性
	// 改正：不是只有叶子节点才会有这个属性，/user和/user/login这两个都有这个属性，这两个路由也都是合法的
	handler HandleFunc

	// 通配符 * 表达的节点，任意匹配
	starChild *node

	// 参数 : 匹配
	paramChild *node
}

// childOf 用于匹配节点
// 查找节点，判断当前的节点的子节点中有没有path节点
// 优先级：精确匹配 > :匹配 > *匹配
// 第一个返回值是匹配到的节点
// 第二个返回值是控制是否是参数匹配：: 和 * 匹配
// 第三个返回值是控制是否匹配到节点
func (n *node) childOf(part string) (*node, bool, bool) {
	// 因为这里是查找，所以不存在当前节点的children属性是nil的情况
	// 只有一种情况会是这样，就是叶子节点
	child, ok := n.children[part]
	if !ok {
		// 如果精确匹配没有匹配到，先用 : 节点匹配，再用 * 节点匹配
		if n.paramChild != nil {
			return n.paramChild, true, n.paramChild != nil
		}
		// 如果精确匹配没有匹配到，就用 * 匹配
		return n.starChild, true, n.starChild != nil
	}
	return child, false, ok
}

// childOrCreate 用于注册路由使用
// 查找节点，判断当前节点的子节点中是否存在path节点，已存在返回path节点，不存在就创建节点并添加到子节点中
func (n *node) childOrCreate(part string) *node {
	if strings.HasPrefix(part, ":") {
		// 是参数 : 的情况
		if n.paramChild == nil { // 多判断一层，如果paramChild不是nil，就表示之前这个路由被注册过了
			n.paramChild = &node{part: part}
		}
		return n.paramChild
	}
	if strings.HasPrefix(part, "*") {
		// 是通配符 * 的情况
		if n.starChild == nil { // 多一层判断，如果starChild不是nil，就表示之前这个路由被注册过了
			n.starChild = &node{part: part}
		}
		return n.starChild
	}
	// 判断当前节点的子节点属性是否为nil
	// 为nil就创建
	if n.children == nil {
		n.children = map[string]*node{}
	}
	child, ok := n.children[part]
	if !ok {
		// part节点不存在，直接创建并添加
		child = &node{
			part: part,
		}
		n.children[part] = child
	}
	return child
}
