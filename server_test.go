package main

import (
	"fmt"
	"log"
	"testing"
)

type V struct {
	Val int `json:"val"`
}

func TestHTTPServer(t *testing.T) {
	server := NewHTTPServer("test")
	server.Get("/query", func(ctx *Context) {
		fmt.Println(ctx.QueryValue("a"))
		fmt.Println(ctx.QueryValue("c"))

		ctx.ResponseWithString(200, "get query val")
	})

	server.Post("/form", func(ctx *Context) {
		val, _ := ctx.FormValue("val").ToInt64()
		fmt.Println(val)
		fmt.Println(ctx.FormValue("val"))

		ctx.ResponseWithString(200, "post form val")
	})
	server.Post("/", func(ctx *Context) {
		a := new(V)
		err := ctx.BindJSON(&a)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(a)

		ctx.ResponseWithString(200, "post json data")
	})

	server.PUT("/", func(ctx *Context) {
		fmt.Println(123)
		ctx.ResponseWithString(200, "put 123")
	})

	server.Delete("/", func(ctx *Context) {
		fmt.Println(123)
		ctx.ResponseWithString(200, "delete 123")
	})

	err := server.Start("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
