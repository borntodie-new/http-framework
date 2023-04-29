package prometheus

import "github.com/borntodie-new/geek-web"

// 可观测性之性能指标

// 这里涉及Prometheus的使用，就不再多说了

type MiddlewarePrometheus struct {
	// 其中可以抽象出一些属性，让用户提供
}

func (m *MiddlewarePrometheus) Builder() geek_web.Middleware {
	return func(next geek_web.HandleFunc) geek_web.HandleFunc {
		return func(ctx *geek_web.Context) {
			next(ctx)
		}
	}
}

func Builder() *MiddlewarePrometheus {
	// 在这里接收用户性的数据，做一些特定的配置信息，返回一个定制化的MiddlewareOpenTelemetry对象实例
	return &MiddlewarePrometheus{}
}
