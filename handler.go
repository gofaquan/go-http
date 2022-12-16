package main

type Handler interface {
	ServeHTTP(c *Context)
	Routable
}

type HandleFunc func(c *Context)
