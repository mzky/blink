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
	winForm.LoadURL("http://127.0.0.1" + port + "/web/index.html")
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
