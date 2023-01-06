package main

type Handler interface {
	serve(c *Context)
	Routable
}

type handleFunc func(c *Context)
