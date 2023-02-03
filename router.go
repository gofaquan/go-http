package main

import (
	"fmt"
	"strings"
)

type router struct {
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

func (r *router) validatePattern(path string) {
	if path == "" {
		panic("http: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("http: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("http: 路由不能以 / 结尾")
	}
}

/*
addRoute 注册路由。
- 已经注册了的路由，无法被覆盖 eg: /user/home 注册两次，会冲突
- path 必须以 / 开始并且结尾不能有 / ,中间也不允许有连续的 /
- 不能在同一个位置注册不同的参数路由 eg: /user/:id 和 /user/:name 冲突
- 不能在同一个位置同时注册通配符路由和参数路由，eg: /user/:id 和 /user/* 冲突
- 同名路径参数，在路由匹配的时候，值会被覆盖。eg: /user/:id/abc/:id，那么 /user/123/abc/456 最终 id = 456
*/
func (r *router) addRoute(method string, path string, handler HandleFunc) {
	if path == "" {
		panic("http: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("http: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("http: 路由不能以 / 结尾")
	}

	root, ok := r.trees[method]
	// 新 HTTP method, 创建根节点
	if !ok {
		// 创建根节点
		root = &node{path: "/"}
		r.trees[method] = root
	}
	if path == "/" {
		if root.handler != nil {
			panic("http: 路由冲突[/]")
		}
		root.handler = handler
		return
	}

	segs := strings.Split(path[1:], "/")
	// 开始一段段处理
	for _, s := range segs {
		if s == "" {
			panic(fmt.Sprintf("http: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		root = root.childOrCreate(s)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("http: 路由冲突[%s]", path))
	}
	root.handler = handler
	root.route = path
}

// findRoute 查找对应节点
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{n: root}, true
	}

	segs := strings.Split(strings.Trim(path, "/"), "/")
	mi := &matchInfo{}
	for _, s := range segs {
		var matchParam bool
		root, matchParam, ok = root.childOf(s)
		if !ok {
			return nil, false
		}
		if matchParam {
			mi.addValue(root.path[1:], s)
		}
	}
	mi.n = root
	return mi, true
}

// node 代表路由树的节点
type node struct {
	path string
	// children 子节点
	// 子节点的 path => node
	children map[string]*node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc
	// route 到达该节点的完整的路由路径
	route string
	// 注册在该节点上的 middleware
	ms []Middleware
	// 通配符 * 表达的节点，任意匹配
	starChild *node
	// 动态参数节点
	paramChild *node

	matchedMs []Middleware
}

func (n *node) childOf(path string) (node *node, isSpecific bool, isMatched bool) {
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	res, ok := n.children[path]
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return res, false, ok
}

// childOrCreate 查找/创建 子节点
func (n *node) childOrCreate(path string) *node {
	if path == "*" { // 首先会判断 path 是不是通配符路径
		if n.paramChild != nil {
			panic(fmt.Sprintf("http: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{path: path}
		}
		return n.starChild
	}

	// 以 : 开头，我们认为是参数路由
	if path[0] == ':' {
		if n.starChild != nil {
			panic(fmt.Sprintf("http: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.paramChild != nil {
			if n.paramChild.path != path {
				panic(fmt.Sprintf("http: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
			}
		} else {
			n.paramChild = &node{path: path}
		}
		return n.paramChild
	}
	// 最后会从 children 里面查找
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		// 如果没有找到，那么会创建一个新的节点
		child = &node{path: path}
		n.children[path] = child
	}
	return child
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		m.pathParams = map[string]string{}
	}
	m.pathParams[key] = value
}
