package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var signals = []os.Signal{
	os.Interrupt, os.Kill, syscall.SIGKILL,
	syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP,
	syscall.SIGABRT, syscall.SIGTERM,
}

type Option func(*App)

type ShutdownCallback func(ctx context.Context)

func WithShutdownCallbacks(cbs ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = append(app.cbs, cbs...)
	}
}

type App struct {
	servers []*HTTPServer

	// 优雅退出整个超时时间，默认30秒
	shutdownTimeout time.Duration

	// 优雅退出时候等待处理已有请求时间，默认10秒
	waitTime time.Duration
	// 自定义回调超时时间，默认三秒钟
	cbTimeout time.Duration

	cbs []ShutdownCallback
}

func NewApp(servers []*HTTPServer, opts ...Option) *App {
	res := &App{
		waitTime:        10 * time.Second,
		cbTimeout:       3 * time.Second,
		shutdownTimeout: 30 * time.Second,
		servers:         servers,
	}
	for _, opt := range opts {
		opt(res)
	}

	return res
}

func (app *App) StartAndServe() {
	for _, s := range app.servers {
		srv := s
		go func() {
			if err := srv.Start(); err != nil {
				if err == http.ErrServerClosed {
					log.Printf("服务器%s已关闭", srv.name)
				} else {
					log.Printf("服务器%s异常退出", srv.name)
				}

			}
		}()
	}

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, signals...)
	<-ch

	go func() {
		select {
		case <-ch:
			log.Printf("强制退出")
			os.Exit(1)
		case <-time.After(app.shutdownTimeout):
			log.Printf("超时强制退出")
			os.Exit(1)
		}
	}()
	app.shutdown()
}

func (app *App) shutdown() {
	log.Println("开始关闭应用，停止接收新请求")
	for _, s := range app.servers {
		s.rejectReq()
	}

	//TODO go func 实时监听请求是否完成
	log.Println("等待正在执行请求完结")
	time.Sleep(app.waitTime)

	log.Println("开始关闭服务器")
	var wg sync.WaitGroup
	wg.Add(len(app.servers))
	for _, srv := range app.servers {
		srvCp := srv
		go func() {
			if err := srvCp.stop(context.Background()); err != nil {
				log.Printf("关闭服务失败%s \n", srvCp.name)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	log.Println("开始执行自定义回调")
	// 执行回调
	wg.Add(len(app.cbs))
	for _, cb := range app.cbs {
		c := cb
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), app.cbTimeout)
			c(ctx)
			cancel()
			wg.Done()
		}()
	}
	wg.Wait()
	// 释放资源
	log.Println("开始释放资源")
	app.close()
}

func (app *App) close() {
	// 释放掉一些可能的资源
	time.Sleep(time.Second)
	log.Println("应用关闭")
}
