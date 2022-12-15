package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, i am going to building a web frame !")
}
func main() {
	http.HandleFunc("/", handler)
	fmt.Println("hello go-http")
	http.ListenAndServe("localhost:8080", nil)
}

// curl localhost:8080
// Hi there, i am going to building a web frame
