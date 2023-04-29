package geek_web

import "fmt"

// Middleware 中间件函数签名
type Middleware func(next HandleFunc) HandleFunc

/*
思考
对于中间件，我们需要实现的方式如下
1. type Middleware func(next HandleFunc) HandleFunc
这是中间件的函数签名
参数next是下一个要执行的中间件
返回值是当前要执行的中间件
*/

/*
下面就是我们想要执行的中间件流程
但是看起来太不优雅了，而且阔扩展性太差
func m() {
	fmt.Println("coming middleware1...")
	func() {
		fmt.Println("coming middleware2...")
		func() {
			fmt.Println("coming middleware3...")
			func() {
				fmt.Println("coming middleware4...")
				func() {
					fmt.Println("coming middleware5...")
					func() {

					}()
					fmt.Println("outing middleware5...")
				}()
				fmt.Println("outing middleware4...")
			}()
			fmt.Println("outing middleware3...")
		}()
		fmt.Println("outing middleware2...")
	}()
	fmt.Println("outing middleware1...")
}
*/
func m() {
	fmt.Println("coming middleware1...")
	func() {
		fmt.Println("coming middleware2...")
		func() {
			fmt.Println("coming middleware3...")
			func() {
				fmt.Println("coming middleware4...")
				func() {
					fmt.Println("coming middleware5...")
					func() {

					}()
					fmt.Println("outing middleware5...")
				}()
				fmt.Println("outing middleware4...")
			}()
			fmt.Println("outing middleware3...")
		}()
		fmt.Println("outing middleware2...")
	}()
	fmt.Println("outing middleware1...")
}
