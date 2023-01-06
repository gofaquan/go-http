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

// sdkHttpServer 这个是基于 net/http 这个包实现的 http server
type sdkHttpServer struct {
	// Name server 的名字，给个标记，日志输出的时候用得上
	Name    string
	handler Handler
	root    Filter
}

func (s *sdkHttpServer) addRoute(method string, pattern string,
	handlerFunc handleFunc) {
	s.handler.addRoute(method, pattern, handlerFunc)
}

func (s *sdkHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r)
	s.root(c)
}
func (s *sdkHttpServer) Start(address string) error {

	return http.ListenAndServe(address, s)
	//return http.ListenAndServe(address, s.handler)
}
func (s *sdkHttpServer) Shutdown(ctx context.Context) error {
	// sleep 一下来模拟这个过程
	fmt.Printf("%s shutdown...\n", s.Name)
	time.Sleep(time.Second)
	fmt.Printf("%s shutdown!!!\n", s.Name)
	return nil
}
func NewSdkHttpServer(name string, builders ...FilterBuilder) *sdkHttpServer {
	handler := NewHandlerBasedOnTree()
	var root Filter = handler.serve
	for i := len(builders) - 1; i >= 0; i-- {
		b := builders[i]
		root = b(root)
	}

	res := &sdkHttpServer{
		Name:    name,
		handler: handler,
		root:    root,
	}
	return res
}

func (s *sdkHttpServer) Get(path string, handler handleFunc) {
	s.handler.addRoute(http.MethodGet, path, handler)
}

func (s *sdkHttpServer) Post(path string, handler handleFunc) {
	s.handler.addRoute(http.MethodPost, path, handler)
}

func (s *sdkHttpServer) Delete(path string, handler handleFunc) {
	s.handler.addRoute(http.MethodDelete, path, handler)
}
func (s *sdkHttpServer) PUT(path string, handler handleFunc) {
	s.handler.addRoute(http.MethodPut, path, handler)
}
