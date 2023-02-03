package main

import (
	"github.com/gofaquan/go-http/middleware/ratelimit"
	"time"
)

type Middleware func(next HandleFunc) HandleFunc

func RateLimit(d time.Duration, capacity int64, opts ...ratelimit.TbOption) Middleware {
	b, err := ratelimit.NewBucket(d, capacity)
	if err != nil {
		panic(err)
	}
	for _, opt := range opts {
		opt(b)
	}
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			b.Wait(1)
			next(ctx)
		}
	}
}
