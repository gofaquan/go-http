package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter

	StatusCode   int
	ResponseData []byte
	PathParams   map[string]string

	MatchedRoute string

	cacheQueryValues url.Values
}

type HandleFunc func(ctx *Context)

func (c *Context) BindJSON(val any) error {
	if c.Request.Body == nil {
		return errors.New("request body 为 nil")
	}

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields() // 出现无法解析的字段报错
	return decoder.Decode(val)
}

type StringVal struct {
	val string
	err error
}

func (c *Context) FormValue(key string) StringVal {
	if err := c.Request.ParseForm(); err != nil {
		return StringVal{err: err}
	}

	return StringVal{val: c.Request.FormValue(key)}
}

func (c *Context) QueryValue(key string) StringVal {
	if c.cacheQueryValues == nil {
		c.cacheQueryValues = c.Request.URL.Query()
	}

	valSlice, ok := c.cacheQueryValues[key]
	if !ok {
		return StringVal{err: errors.New("找不到对应 key")}
	}
	return StringVal{val: valSlice[0]}
}

func (c *Context) PathValue(key string) StringVal {
	val, ok := c.PathParams[key]
	if !ok {
		return StringVal{err: errors.New("找不到对应 key")}
	}

	return StringVal{val: val}
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.ResponseWriter, cookie)
}

func (c *Context) ResponseWithString(code int, msg string) error {
	c.ResponseData = []byte(msg)
	c.StatusCode = code
	return nil
}

func (c *Context) ResponseWithJSON(code int, val any) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	c.ResponseWriter.Header().Set("Content-Type", "application/json")
	c.StatusCode = code
	c.ResponseData = bytes
	return err
}

func (s StringVal) String() (string, error) {
	return s.val, s.err
}
func (s StringVal) ToInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}
func (s StringVal) ToUInt64() (uint64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseUint(s.val, 10, 64)
}

func (c *Context) StatusOK(msg string) error {
	return c.ResponseWithString(http.StatusOK, msg)
}
func (c *Context) StatusNotFound(msg string) error {
	return c.ResponseWithString(http.StatusNotFound, msg)
}
func (c *Context) StatusInternalServerError(msg string) error {
	return c.ResponseWithString(http.StatusInternalServerError, msg)
}
func (c *Context) JSONStatusOK(val any) error {
	return c.ResponseWithJSON(http.StatusOK, val)
}
func (c *Context) JSONStatusNotFound(val any) error {
	return c.ResponseWithJSON(http.StatusNotFound, val)
}
func (c *Context) JSONStatusInternalServerError(val any) error {
	return c.ResponseWithJSON(http.StatusInternalServerError, val)
}
