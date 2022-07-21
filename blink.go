package blink

//#include "blink.h"
import "C"

import (
	"fmt"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/lxn/win"
	"github.com/mzky/blink/internal/devtools"
	"github.com/mzky/blink/internal/dll"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

//任务队列,保证所有的API调用都在痛一个线程
var jobQueue = make(chan func())

//初始化blink,释放并加载dll,启动调用队列
func InitBlink() error {
	//准备释放dll到临时目录
	dllByte, err := dll.Asset("blink.dll")
	if err != nil {
		return fmt.Errorf("找不到内嵌dll,err: %+v", err)
	}

	tmpFile, err := ioutil.TempFile(TempPath, "blink_*.dll")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %+v", err)
	}
	dllPath := tmpFile.Name()
	if err := ioutil.WriteFile(dllPath, dllByte, 755); err != nil {
		return fmt.Errorf("解出dll文件失败: %+v", err)
	}
	// 这里不能用defer
	tmpFile.Close()

	//启动一个新的协程来处理blink的API调用
	go func() {
		//将这个协程锁在当前的线程上
		runtime.LockOSThread()

		//初始化
		C.initBlink(
			C.CString(dllPath),
			C.CString(TempPath),
			C.CString(filepath.Join(TempPath, "cookie.dat")),
		)

		//注册DevTools工具到虚拟文件系统
		RegisterFileSystem("__devtools__", &assetfs.AssetFS{
			Asset:     devtools.Asset,
			AssetDir:  devtools.AssetDir,
			AssetInfo: devtools.AssetInfo,
		})

		//消费API调用,同时处理好windows消息
		for {
			select {
			case job := <-jobQueue:
				job()
			default:
				//消息循环
				msg := &win.MSG{}
				if win.GetMessage(msg, 0, 0, 0) != 0 {
					win.TranslateMessage(msg)
					//是否传递下去
					next := true
					//拿到对应的webview
					view := getWebViewByHandle(msg.HWnd)
					if view != nil {
						next = view.processMessage(msg)
					}
					if next {
						win.DispatchMessage(msg)
					}
				}
			}
		}
	}()

	logger.Println("blink初始化完毕")

	return nil
}
