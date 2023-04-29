package geek_web

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// 文件处理这块，其实包含三个内容
// 1. 文件的上传
// 2. 文件的下载
// 3. 静态文件的读取

// 文件上传和文件下载没什么好说的，大家自行百度、Google即可，网上一大把的
// 上传和下载的功能建议能用现成的云服务就用云服务，性能和安全上都很可靠稳定

// 这里我们只实现一个静态文件功能，因为这个功能可以和页面渲染很好的搭配在一起

// StaticFileHandler 静态文件
type StaticFileHandler struct {
	// openPath 开放本地的文件夹
	openPath string

	// 下面两个属性功能如下
	// /assets/:filepath
	// prefix = assets
	// paramsKey = filepath
	// Prefix 路由前缀前缀
	Prefix string
	// ParamsKey 参数map的key
	ParamsKey string

	// cache 缓存住静态文件
	cache *lru.Cache[string, interface{}]
	// maxCacheFileCnt 最大能够缓存多少静态文件
	maxCacheFileCnt int
	// fileMaxSize 每个文件的最大长度
	//perFileSize int
}

// StaticFileHandlerOpt 由于缓存不是每个用户都需要的，所以这里做成有个可选项
type StaticFileHandlerOpt func(handler *StaticFileHandler)

// StaticFileWithCache 为StaticFileHandler配置一个缓存队列
// maxCacheFileCnt 最多能缓存这么多个数据
// perFileSize 缓存文件的最大值
func StaticFileWithCache(maxCacheFileCnt int, perFileSize int) StaticFileHandlerOpt {
	return func(handler *StaticFileHandler) {
		// 初始化一个cache对象
		// 初始化的方法方法有点奇怪，如果之前没接触过泛型的话
		// string是作为缓存的key的类型
		// interface{}是作为缓存value的类型
		// 传入的参数是缓存最多能缓存多少数据
		c, _ := lru.New[string, interface{}](maxCacheFileCnt * perFileSize)
		handler.maxCacheFileCnt = maxCacheFileCnt
		handler.cache = c
	}
}

func NewStaticFileHandler(openPath, prefix, paramsKey string, opts ...StaticFileHandlerOpt) *StaticFileHandler {
	s := &StaticFileHandler{
		openPath:  openPath,
		Prefix:    prefix,
		ParamsKey: paramsKey,
	}
	// 重点是这里
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *StaticFileHandler) Handler(ctx *Context) {
	// 1. 拿到文件名
	fileName, _ := ctx.Param(s.ParamsKey)
	// 2. 读取文件
	// 2.1 如果用户传入一个不存在的文件名，最终文件肯定是读取不到的
	// 2.2 如果用户知道我们的文件结构，传入一些我们的系统文件名，那怎么办？
	filePath := filepath.Join(s.openPath, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetData([]byte("Server Internal Error, Please Try Again Later!"))
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetData([]byte("Server Internal Error, Please Try Again Later!"))
		return
	}
	// 3. 写入数据到响应中
	ctx.SetStatusCode(http.StatusOK)
	ctx.SetData(data)
}
