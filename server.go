package main

import (
	"log"
	"net"
	"net/http"
	"strconv"
)

type Server interface {
	http.Handler
	Start(address string) error
	addRoute(method, path string, handler HandleFunc, ms ...Middleware)
}

type HTTPServer struct {
	router
	log *log.Logger
}

type ServerOption func(server *HTTPServer)

func NewHTTPServer(opts ...ServerOption) *HTTPServer {
	s := &HTTPServer{
		router: newRouter(),
		log:    log.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Request:        r,
		ResponseWriter: w,
	}

	h.serve(ctx)
}

func (h *HTTPServer) serve(ctx *Context) {
	target, ok := h.findRoute(ctx.Request.Method, ctx.Request.URL.Path)
	if target.n != nil {
		ctx.PathParams = target.pathParams
		ctx.MatchedRoute = target.n.route
	}

	var root HandleFunc = func(ctx *Context) {
		if !ok || target.n == nil || target.n.handler == nil {
			ctx.StatusCode = 404
			ctx.ResponseData = []byte("404 NOT FOUND")
			return
		}
		target.n.handler(ctx)
	}

	for i := len(target.mdls) - 1; i >= 0; i-- {
		root = target.mdls[i](root)
	}

	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flushResponse(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

func (h *HTTPServer) flushResponse(ctx *Context) {
	if ctx.StatusCode > 0 {
		ctx.ResponseWriter.WriteHeader(ctx.StatusCode)
	}
	ctx.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(len(ctx.ResponseData)))
	_, err := ctx.ResponseWriter.Write(ctx.ResponseData)
	if err != nil {
		h.log.Fatalln("回写响应失败", err)
	}
}

func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return http.Serve(l, h)
	//return http.ListenAndServe(addr, h)
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}
func (h *HTTPServer) Delete(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodDelete, path, handleFunc)
}

func (h *HTTPServer) PUT(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPut, path, handleFunc)
}
func (h *HTTPServer) Options(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodOptions, path, handleFunc)
}
