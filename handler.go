package main

type Handler interface {
	ServerHTTP(c *Context)
	Routable
}

type HandleFunc func(c *Context)
