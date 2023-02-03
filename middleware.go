package main

type Middleware func(next HandleFunc) HandleFunc
