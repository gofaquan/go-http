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

// linux signal
var signals = []os.Signal{
	//os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGSTOP,
	syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP,
	//syscall.SIGABRT, syscall.SIGSYS, syscall.SIGTERM,
}

type Option func(*App)

type ShutdownCallback func(ctx context.Context)

func WithShutdownCallbacks(cbs ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = cbs
	}
}

type App struct {
	servers []*Server

	// 优雅退出整个超时时间，默认30秒
	shutdownTimeout time.Duration

	// 优雅退出时候等待处理已有请求时间，默认10秒钟
	waitTime time.Duration
	// 自定义回调超时时间，默认三秒钟
	cbTimeout time.Duration

	cbs []ShutdownCallback
}

func NewApp(servers []*Server, opts ...Option) *App {
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
	println("hello")
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

	log.Println("等待正在执行请求完结")
	// 这里可以改造为实时统计正在处理的请求数量，为0 则下一步
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
	// 在这里释放掉一些可能的资源
	time.Sleep(time.Second)
	log.Println("应用关闭")
}

type Server struct {
	srv  *http.Server
	name string
	mux  *serverMux
}

type serverMux struct {
	reject bool
	*http.ServeMux
}

func (s *serverMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.reject {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("服务已关闭"))
		return
	}
	s.ServeMux.ServeHTTP(w, r)
}

func NewServer(name string, addr string) *Server {
	mux := &serverMux{ServeMux: http.NewServeMux()}
	return &Server{
		name: name,
		mux:  mux,
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) rejectReq() {
	s.mux.reject = true
}

func (s *Server) stop(ctx context.Context) error {
	log.Printf("服务器 %s 关闭中", s.name)
	return s.srv.Shutdown(ctx)
}
