package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestMapHandler_Route(t *testing.T) {
	handler := NewMapHandler()
	handler.Route(http.MethodGet, "/user", func(c *Context) {})
	_, ok := handler.handlers.Load("GET#/user")
	// 判断是否有这么一个路由
	assert.Equal(t, ok, true)
}
