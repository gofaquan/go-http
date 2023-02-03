package main

//
//func handler(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprintf(w, "Hi there, i am going to building a web frame !")
//}
//func main() {
//	//http.HandleFunc("/", handler)
//	//fmt.Println("hello go-http")
//	//http.ListenAndServe("localhost:8080", nil)
//
//	TestShutdown2()
//}
//
//// curl localhost:8080
//// Hi there, i am going to building a web frame
//
//// 注意要从命令行启动，否则不同的 IDE 可能会吞掉关闭信号
//func TestShutdown2() {
//	s1 := NewSdkHttpServer("cnm")
//	s1.addRoute("GET", "/a", func(c *Context) {
//		_, _ = c.Writer.Write([]byte("hello"))
//	})
//
//	err := s1.Start("8080")
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	//app := service.NewApp([]*service.Server{s1, s2}, service.WithShutdownCallbacks(StoreCacheToDBCallback))
//	//app.StartAndServe()
//}
