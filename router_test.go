package geek_web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddRouter(t *testing.T) {
	mockHandler := func(ctx *Context) {}

	testRouter := []struct {
		name    string
		method  string
		pattern string
	}{
		{

			name:    "测试 GET /user",
			method:  "GET",
			pattern: "/user",
		},
		{
			name:    "测试 POST /order",
			method:  "POST",
			pattern: "/user",
		},
		{
			name:    "测试 GET /user/login",
			method:  "GET",
			pattern: "/user/login",
		},
		{
			name:    "测试 GET /",
			method:  "GET",
			pattern: "/",
		},
		{
			name:    "测试 POST /",
			method:  "POST",
			pattern: "/",
		},
		{
			name:    "错误 GET //user/home",
			method:  "GET",
			pattern: "//user/home",
		},
		{
			name:    "错误 GET book/info",
			method:  "GET",
			pattern: "book/info",
		},
		{
			name:    "错误 GET /hero/id/",
			method:  "GET",
			pattern: "/hero/id/",
		},
	}

	r := newRouter()
	for _, tt := range testRouter {
		t.Run(tt.name, func(t *testing.T) {
			r.addRouter(tt.method, tt.pattern, mockHandler)
		})
		// r.addRouter(tt.method, tt.pattern, mockHandler)
	}
}

func TestStarAddRouter(t *testing.T) {
	mockHandler := func(ctx *Context) {}

	testRouter := []struct {
		name    string
		method  string
		pattern string
	}{
		{

			name:    "测试 GET /asserts/*filepath",
			method:  "GET",
			pattern: "/asserts/*filepath",
		},
		{
			name:    "测试 GET /asserts/*filepath",
			method:  "GET",
			pattern: "/asserts/*filepath",
		},
	}

	r := newRouter()
	for _, tt := range testRouter {
		//t.Run(tt.name, func(t *testing.T) {
		//	r.addRouter(tt.method, tt.pattern, mockHandler)
		//})
		r.addRouter(tt.method, tt.pattern, mockHandler)
	}
	t.Log(r)
}

func TestStarFindRouter(t *testing.T) {
	mockHandler := func(ctx *Context) {}

	testRouter := []struct {
		name    string
		method  string
		pattern string
	}{
		{

			name:    "测试 GET /asserts/*filepath",
			method:  "GET",
			pattern: "/asserts/*filepath",
		},
		//{
		//	name:    "测试 GET /asserts/*filepath",
		//	method:  "GET",
		//	pattern: "/asserts/*filepath",
		//},
	}

	r := newRouter()
	for _, tt := range testRouter {
		r.addRouter(tt.method, tt.pattern, mockHandler)
	}
	_, params, ok := r.findRouter("GET", "/asserts/css/neo.css/ausdhwd/asfudif")
	assert.True(t, ok)
	t.Log(params)
}
