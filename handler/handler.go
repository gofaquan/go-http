package handler

import (
	"github.com/gofaquan/go-http/context"
	"github.com/gofaquan/go-http/server"
)

type Handler interface {
	ServerHTTP(c *context.Context)
	server.Routable
}

type HandleFunc func(c *context.Context)
