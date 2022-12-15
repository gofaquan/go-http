package server

import "github.com/gofaquan/go-http/handler"

// Server 定义 http Server 的顶级抽象
type Server interface {
	Start(address string) error
	Routable
}

// Routable 可路由的
type Routable interface {
	// Route 设定一个路由，命中该路由的会执行 handleFunc 的代码
	Route(method, pattern string, handleFunc handler.HandleFunc)
}
