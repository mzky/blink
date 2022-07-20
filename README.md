# blink
使用html来编写golang的GUI程序(only windows), 基于[miniblink开源库](https://github.com/weolar/miniblink49)  

## Demo
[Demo项目地址](https://github.com/raintean/blink-demo)

## 特性
---
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
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/mzky/blink"
	"io/fs"
	"log"
	"net"
	"net/http"
	"pcmClient/utils"
	"strconv"
	"time"
)

//go:embed res
var res embed.FS

var (
	c       utils.CorsVar
	mainForm *blink.WebView
)

/*
减肥并隐藏控制台
go build -ldflags "-w -s -H=windowsgui"
debug模式，F12可以调出开发者工具
go build -ldflags "-w -s -H=windowsgui" -tags="bdebug"
*/
func main() {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	c.LocalPort = strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	c.SrvPort = ":7569"

	r := gin.Default()
	//gin.SetMode(gin.ReleaseMode)
	r.Any("/kerb/*path", c.PublicProxy)
	r.GET("/health", c.Health)
	r.POST("/exit", c.Exit)
	r.POST("/setAddress", c.SetAddress)

	fRes, err := fs.Sub(res, "res") // 必须
	if err != nil {
		panic(err)
	}
	r.StaticFS("/web", http.FS(fRes))
	go r.Run(":" + c.LocalPort)
	
	if mainForm == nil {
		// 是否支持F5和F12调试模式
		// blink.SetDebugMode(false)
		err := blink.InitBlink()
		if err != nil {
			log.Fatal(err)
		}
		<-time.After(time.Second)
		mainForm = blink.NewWebView(false, 1280, 800)

		mainForm.LoadURL("http://127.0.0.1:" + c.LocalPort + "/web/login.html")
		// 设置窗体标题
		mainForm.SetWindowTitle("升级维护客户端v1.3.1")
		mainForm.DisableAutoTitle()

		icoByte, _ := res.ReadFile("res/gear22.ico")
		mainForm.SetWindowIconFromBytes(icoByte)
		// 移动到屏幕中心位置
		mainForm.MoveToCenter()
		mainForm.ShowWindow()

		//当窗口被销毁的时候,变量=nil
		mainForm.On("destroy", func(_ *blink.WebView) {
			mainForm = nil
		})
		<-make(chan bool)
	} else {
		//窗口实例存在,则提到前台
		mainForm.ToTop()
		mainForm.RestoreWindow()
		mainForm.ShowWindow()
	}
}

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
