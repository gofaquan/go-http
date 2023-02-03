package main

import (
	"context"
	"fmt"

	"net/http"
	"time"
)

// Server 定义 http Server 的顶级抽象
type Server interface {
	Start(address string) error
	Shutdown(ctx context.Context) error
	Routable
}

// Routable 可路由的
type Routable interface {
	// Route 设定一个路由，命中该路由的会执行 handleFunc 的代码
	addRoute(method, pattern string, handleFunc handleFunc)
}

// SdkHttpServer 这个是基于 net/http 这个包实现的 http server
type SdkHttpServer struct {
	// Name server 的名字，给个标记，日志输出的时候用得上
	Name    string
	handler Handler
	root    Filter
}

func (s *SdkHttpServer) addRoute(method string, pattern string,
	handlerFunc handleFunc) {
	s.handler.addRoute(method, pattern, handlerFunc)
}

func (s *SdkHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r) // 新建 context 处理 http 请求
	s.root(c)             // 中间件调用
}

func (s *SdkHttpServer) Start(address string) error {
	return http.ListenAndServe(address, s)
}

func (s *SdkHttpServer) Shutdown(ctx context.Context) error {
	fmt.Printf("%s shutdown...\n", s.Name)
	time.Sleep(time.Second)
	fmt.Printf("%s shutdown!!!\n", s.Name)
	return nil
}

func NewSdkHttpServer(name string, builders ...FilterBuilder) *SdkHttpServer {
	handler := NewHandlerBasedOnTree()
	var root Filter = handler.serve // 最外面的中间就执行路由的 func
	// 反着加入责任链
	for i := len(builders) - 1; i >= 0; i-- {
		b := builders[i]
		root = b(root)
	}

	res := &SdkHttpServer{
		Name:    name,
		handler: handler,
		root:    root,
	}

	return res
}

func (s *SdkHttpServer) Get(path string, handler handleFunc) {
	s.handler.addRoute(http.MethodGet, path, handler)
}
func (s *SdkHttpServer) Post(path string, handler handleFunc) {
	s.handler.addRoute(http.MethodPost, path, handler)
}
func (s *SdkHttpServer) Delete(path string, handler handleFunc) {
	s.handler.addRoute(http.MethodDelete, path, handler)
}
func (s *SdkHttpServer) PUT(path string, handler handleFunc) {
	s.handler.addRoute(http.MethodPut, path, handler)
}
