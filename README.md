# blink
- fork自https://github.com/raintean/blink
- 使用html来编写golang的GUI程序(only windows), 基于[miniblink开源库](https://github.com/weolar/miniblink49)  
- 依赖cgo

## 变更部分 (This)
- dll释放时通过os.CreateTemp随机命名(以blink_开头)
- dll更新到20220405
- 更新示例，go:embed方式打包web页面，gin框架使用
- 增加弹窗接口MessageBox(标题，内容)
- 增加单实例锁、查找并置顶窗口等
- dll做upx压缩，dll打包的代码，用go:embed替代（32位没有环境试）
- 此库对页面下载没有响应，需要下载功能可切换到`https://github.com/mzky/goblink`（这个变量命名让人沉醉，我正视图改进）

## Demo
[Demo项目地址](https://github.com/raintean/blink-demo)

## 特性
- [x] 一个可执行文件, miniblink的dll被嵌入库中
- [x] 生成的可执行文件灰常小
- [x] 支持无缝golang和浏览器页面js的交互 (Date类型都做了处理), 并支持异步调用golang中的方法(Promise), 支持异常获取.
- [x] 嵌入开发者工具，bdebug构建tags开启 ( <b>go build -ldflags "-w -s -H=windowsgui" -tags="bdebug"</b> )
- [x] 支持虚拟文件系统, 基于golang的http.FileSystem, 意味着go-bindata出的资源可以直接嵌入程序, 无需开启额外的http服务
- [x] 添加了部分简单的接口(最大化,最小化,无任务栏图标等)
- [x] 设置窗口图标(参见icon.go文件)
- [ ] 支持文件拖拽
- [ ] 自定义dll,而不是使用内嵌的dll(防止更新不及时)
- [ ] golang调用js方法时的异步.
- [ ] dll的内存加载, 尝试过基于MemoryModule的方案, 没有成功, 目前是释放dll到临时目录, 再加载.
- [ ] 还有很多...

## 安装
```bash
go get github.com/mzky/blink
```

## 示例
```go
package main

import (
	"context"
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/mzky/blink"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed res
var res embed.FS

var (
	winForm     *blink.WebView
	windowTitle = "blink_example"
	port        = ":7569"
)

func init() {
	// 是否支持F5和F12调试模式
	// blink.SetDebugMode(false)
	if err := blink.InitBlink(); err != nil {
		log.Fatal(err)
	}
	winForm = blink.NewWebView(false, 1280, 800)
	if err := winForm.LockMutex("PCM_Client"); err != nil {
		winForm.MessageBox("消息", "同时只能运行一个升级维护客户端！")
		winForm.FindWindowToTop(windowTitle)
		<-time.After(time.Second)
		log.Fatalf("客户端已经运行 %+v", err)
	}
}

func main() {
	srv := &http.Server{Addr: port, Handler: initHandler()}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			winForm.MessageBox("消息", "本地7569端口被占用，关闭占用7569端口程序后重试")
			log.Fatalf("端口被占用:%+v", err)
		}
	}()

	<-time.After(time.Second)
	winForm.LoadURL("http://127.0.0.1" + port + "/web/login.html")
	// 设置窗体标题
	winForm.SetWindowTitle(windowTitle)
	winForm.DisableAutoTitle()

	icoByte, _ := res.ReadFile("res/gear22.ico")
	winForm.SetWindowIconFromBytes(icoByte)
	// 移动到屏幕中心位置
	winForm.MoveToCenter()
	winForm.ShowWindow()
	winForm.ToTop()

	//当窗口被销毁的时候,变量=nil
	winForm.On("destroy", func(_ *blink.WebView) {
		winForm = nil
		log.Println("Shutdown Service ...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 关闭时退出httpserver，否则进程一直还在
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
		os.Exit(0)
	})
	<-make(chan bool)

}

func initHandler() *gin.Engine {
	r := gin.Default()
	//gin.SetMode(gin.ReleaseMode)

	fRes, err := fs.Sub(res, "res") // 必须
	if err != nil {
		panic(err)
	}
	r.StaticFS("/web", http.FS(fRes))

	return r
}

/*
-- 减肥并隐藏控制台
go build -ldflags "-w -s -H=windowsgui"
-- debug模式，F12可以调出开发者工具
go build -ldflags "-w -s -H=windowsgui" -tags="bdebug"
*/
```

## golang和js交互
js调用/获取golang中的方法或者值,异常可捕获
> main.go
```golang
//golang注入方法
view.Inject("GetData", func(num int) (int, error) {
	if num > 10 {
		return 0, errors.New("num不能大于10")
	} else {
		return num + 1, nil
	}
})

//golang注入值
view.Inject("Data", "a string")
```
> index.js
```javascript
await BlinkFunc.GetData(10) //-> 11
await BlinkFunc.GetData(11) //-> throw Error("num不能大于10")
BlinkData.Data // -> "a string"
```
golang调用/获取javascript中的方法或者值,异常可捕获(err变量返回)
> index.js
```javascript
window.Foo = new Date();
window.Bar = function (name) {
    return `hello ${name}`;
};
```
> main.go
```golang
value, err := view.Invoke("Foo")
value.ToXXX // -> Time(golang类型)
value, err := view.Invoke("Bar", "blink")
value.ToString() // -> "hello blink"
```
## 注意
- 网页调试工具默认不打包进可执行文件,请启用BuildTags **bdebug**, eg. `go build -tags bdebug`
- 使用本库需依赖cgo编译环境(mingw32)

## ...
再次感谢miniblink项目, 另外如果觉得本项目好用请点个星.  
欢迎PR, > o <
