package main

import (
	"net/http"
	"strings"
)

type node struct {
	path     string
	children []*node
	handler  HandleFunc
}

func newNode(path string) *node {
	return &node{path: path, children: make([]*node, 0, 2)}
}

type TreeHandler struct {
	root *node
}

func (t *TreeHandler) ServeHTTP(c *Context) {
	// 查找 router tree
	handler, found := t.search(c.Request.URL.Path)
	if !found { // 不存在对于路由
		c.Writer.WriteHeader(http.StatusNotFound)
		c.Writer.Write([]byte("未找到对于路由!"))
		return
	}
	handler(c)
}

func (t *TreeHandler) Route(method, pattern string, handleFunc HandleFunc) {
	// 将 pattern 按照 URL 的分隔符切割
	// 例如，/user/friends 将变成 [user, friends]
	// 将前后的 / 去掉，统一格式
	pattern = strings.Trim(pattern, "/")
	paths := strings.Split(pattern, "/")
	// 当前指向根节点
	cur := t.root

	for index, path := range paths {
		child, found := t.matchChild(cur, path)
		if found {
			cur = child
		} else {
			t.buildSubTree(cur, paths[index:], handleFunc)
			return
		}
	}
	// 离开了循环，说明我们加入的是短路径，
	// 比如说我们先加入了 /order/detail
	// 再加入/order，那么就会执行
	cur.handler = handleFunc
}

func (t *TreeHandler) search(path string) (HandleFunc, bool) {
	// 去除头尾可能有的/，然后按照/切割成段
	paths := strings.Split(strings.Trim(path, "/"), "/")
	cur := t.root
	for _, path := range paths {
		child, found := t.matchChild(cur, path)
		if !found {
			return nil, false
		}
		cur = child
	}
	// 没有加入 handler 返回 false
	if cur.handler == nil {
		// 排除类似 /user/profile 有 handler, 而 /user 没有 handler 的场景
		return nil, false
	}
	return cur.handler, true
}

// MatchChild 从子节点里边找一个匹配到了当前 path 的节点
func (t *TreeHandler) matchChild(root *node, path string) (*node, bool) {
	for _, child := range root.children {
		if child.path == path {
			return child, true
		}
	}
	return nil, false
}

// 找不到子节点就构造子树
func (t *TreeHandler) buildSubTree(root *node, paths []string, handlerFn HandleFunc) {
	cur := root
	for _, path := range paths {
		newNode := newNode(path)
		cur.children = append(cur.children, newNode)
		cur = newNode
	}
	cur.handler = handlerFn
}
