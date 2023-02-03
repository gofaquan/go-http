package main

import (
	"fmt"
	"log"
	"testing"
)

func TestHTTPServer(t *testing.T) {
	server := NewHTTPServer()
	server.Get("/", func(ctx *Context) {
		fmt.Println(123)
		ctx.ResponseWithString(200, "get 1234")
	})

	server.Post("/", func(ctx *Context) {
		fmt.Println(123)
		ctx.ResponseWithString(200, "post 1234")
	})

	server.PUT("/", func(ctx *Context) {
		fmt.Println(123)
		ctx.ResponseWithString(200, "put 1234")
	})

	server.Delete("/", func(ctx *Context) {
		fmt.Println(123)
		ctx.ResponseWithString(200, "delete 1234")
	})

	err := server.Start("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
