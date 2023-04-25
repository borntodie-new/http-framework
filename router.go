package geek_web

// router 路由树的结构
// 1. 提供注册的功能
// 2. 提供匹配的功能
type router struct {
	// trees "路由树"，应该叫路由森林，因为每一个请求方法都会对应一颗树
	// 具体结构：{GET: tree, POST: tree, ...}
	trees map[string]*node
}

// addRouter 注册路由
func (r *router) addRouter(method string, path string, handleFunc HandleFunc) {

}

// findRouter 匹配路由
func (r *router) findRouter(method string, path string) *node {
	return nil
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
func (n *node) childOf(path string) (*node, bool) {
	return nil, false
}

// childOrCreate 用于注册路由使用
// 查找节点，判断当前节点的子节点中是否存在path节点，已存在返回path节点，不存在就创建节点并添加到子节点中
func (n *node) childOrCreate(path string) *node {
	return nil
}
