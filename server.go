package main

import (
	"log"
	"net/http"
)

type Server interface {
	http.Handler
	Start(address string) error
	addRoute(method, path string, handler HandleFunc, ms ...Middleware)
}

type HTTPServer struct {
	router
	log.Logger
}
