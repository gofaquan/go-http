package main

import (
	"fmt"
	"net/http"
	"sync"
)

type MapHandler struct {
	handlers sync.Map
}

func (m *MapHandler) key(method, path string) string {
	// 默认 method 与 path 用一个 # 隔开
	return fmt.Sprintf("%s#%s", method, path)
}

func (m *MapHandler) ServerHTTP(c *Context) {
	request := c.Request
	key := m.key(request.Method, request.URL.Path)
	// 加载路由
	handler, ok := m.handlers.Load(key)
	if !ok {
		c.Writer.WriteHeader(http.StatusNotFound)
		c.Writer.Write([]byte("没有匹配到任何路由!"))
		return
	}
	handler.(HandleFunc)(c)
}

func (m *MapHandler) Route(method, pattern string, handleFunc HandleFunc) {
	key := m.key(method, pattern)
	// 存进 sync map
	m.handlers.Store(key, handleFunc)

}

func NewMapHandler() *MapHandler {
	return &MapHandler{}
}
