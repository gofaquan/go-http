package main

import (
	"encoding/json"
	"io"
	"net/http"
)

// Context 自定义, 包含基本的 write & request
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

func (c *Context) ReadJson(data interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, data)
}

func (c *Context) WriteJson(status int, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.Writer.Write(bytes)
	if err != nil {
		return err
	}
	c.Writer.WriteHeader(status)
	return nil
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
	}
}
