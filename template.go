package geek_web

import (
	"bytes"
	"html/template"
)

// 页面渲染

/*
## 写在前面
页面渲染功能是前后端不分离架构的产物，它的作用是，在服务端在就将全部的前端页面渲染好。接下来就是将这些数据直接返回到浏览器
浏览器会直接识别HTML标签

对于这个功能，其实不算是一个Web框架的核心功能，至少对于现在的前后端分离的项目来说是这样的。

我们对这个功能的有两个思路
1. 完全自己写，框架实现一套模板语言出来——不太现实，模板语言严格来说其实就算是一种语言了
2. 框架提供接口，实现用户提供

我们就还是采用方式二。对比可观测性的实现方案来说，这个需求我们的框架才是真正的提供接口
可观测性方案我们并没有提供在框架层面提供一个接口，而是结合AOP实现可插拔式的集成

这里我们提供一个接口 + 实现一个基于Go内置的模板语言的引擎

模板引擎最终会集成到Context上下文中，在Context的HTML方法中使用
*/

type TemplateEngine interface {
	// Render 渲染数据
	// ctx 上下文对象
	// templateName 文件模板的名称
	// data 动态数据，需要渲染到template里面的数据
	Render(ctx *Context, templateName string, data any) ([]byte, error)

	// ParseGlob 解析模板，耦合度太高了，就不在接口层面限制了
	// ParseGlob(pattern string) error
}

type GoTemplateEngine struct {
	// T Go内置的模板引擎对象
	T *template.Template
}

func (g *GoTemplateEngine) Render(ctx *Context, templateName string, data any) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := g.T.ExecuteTemplate(buf, templateName, data)
	return buf.Bytes(), err
}

// ParseGlob 解析模板
func (g *GoTemplateEngine) ParseGlob(pattern string) error {
	var err error
	g.T, err = template.ParseGlob(pattern)
	return err
}

// NewGoTemplateEngine 实例化一个Go内置的模板引擎
func NewGoTemplateEngine() *GoTemplateEngine {
	return &GoTemplateEngine{}
}
