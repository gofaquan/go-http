package main

import (
	"context"
	"fmt"
	"testing"

	"log"
	"time"
)

// === RUN   TestApp
//2023/02/05 16:29:38 开始关闭应用，停止接收新请求
//2023/02/05 16:29:38 等待正在执行请求完结
//2023/02/05 16:29:48 开始关闭服务器
//2023/02/05 16:29:48 服务器 admin 关闭中
//admin shutdown...
//2023/02/05 16:29:48 服务器 business 关闭中
//business shutdown...
//business shutdown!!!
//admin shutdown!!!
//2023/02/05 16:29:49 开始执行自定义回调
//2023/02/05 16:29:49 刷新缓存中……
//2023/02/05 16:29:52 刷新缓存超时
//2023/02/05 16:29:52 开始释放资源
//2023/02/05 16:29:54 应用关闭
//--- PASS: TestApp (18.06s)
//PASS
//
//
//Process finished with the exit code 0

// 注意要从命令行启动，否则不同的 IDE 可能会吞掉关闭信号
func TestApp_Shutdown(t *testing.T) {
	s1 := NewHTTPServer("business", "localhost:8080") //

	s1.Post("/form", func(ctx *Context) {
		val, _ := ctx.FormValue("val").ToInt64()
		fmt.Println(val)
		fmt.Println(ctx.FormValue("val"))

		ctx.ResponseWithString(200, "post form val")
	})

	s2 := NewHTTPServer("admin", "localhost:8081") //
	s2.Get("/query", func(ctx *Context) {
		fmt.Println(ctx.QueryValue("a"))
		fmt.Println(ctx.QueryValue("c"))

		ctx.ResponseWithString(200, "get query val")
	})

	servers := []*HTTPServer{s1, s2}
	app := NewApp(servers, WithShutdownCallbacks(StoreCacheToDBCallback))
	app.StartAndServe()
}

func StoreCacheToDBCallback(ctx context.Context) {
	done := make(chan struct{}, 1)
	go func() {
		log.Printf("刷新缓存中……")
		time.Sleep(1 * time.Second)
	}()
	select {
	case <-ctx.Done():
		log.Printf("刷新缓存超时")
	case <-done:
		log.Printf("缓存被刷新到了 DB")
	}
}
