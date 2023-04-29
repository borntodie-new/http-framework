package geek_web

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// 异常恢复中间件

// 由于用户的业务逻辑中可能会出现各种意想不到的panic错误
// 对于我们的框架必须支持对这种panic的recover,也就是hook住这些错误

type MiddlewareRecovery struct {
	// 提供一个属性，供用户选择将错误信息输出到何处
	logFunc func(information string)
}

func (m *MiddlewareRecovery) Builder() Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			// 这个defer负责hook住所有的panic错误
			defer func() {
				if err := recover(); err != nil {
					// 下面的是输出给客户端看的
					ctx.SetStatusCode(http.StatusInternalServerError)
					ctx.SetData([]byte("Server Internal Error, Please Try Again Later!"))
					// 下面的输出给开发者看的
					m.logFunc(m.trace(fmt.Sprintf("%s\n", err)))
					return
				}
			}()
			// 执行后续逻辑
			next(ctx)
		}
	}
}

func (m *MiddlewareRecovery) trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func RecoveryBuilder(logFunc func(information string)) *MiddlewareRecovery {
	if logFunc == nil {
		logFunc = func(information string) {
			fmt.Println(information)
		}
	}
	return &MiddlewareRecovery{logFunc: logFunc}
}
