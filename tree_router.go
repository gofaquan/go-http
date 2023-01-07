package main

import (
	"fmt"
	"net/http"
	"strings"
)

//var ErrorInvalidRouterPattern = errors.New("invalid router pattern")

type node struct {
	path     string
	children []*node

	// 如果这是叶子节点，
	// 那么匹配上之后就可以调用该方法
	handler handleFunc
}

func newNode(path string) *node {
	return &node{
		path:     path,
		children: make([]*node, 0, 2),
	}
}

type HandlerBasedOnTree struct {
	methodTrees map[string]*node
}

func NewHandlerBasedOnTree() Handler {
	return &HandlerBasedOnTree{
		methodTrees: map[string]*node{},
	}
}

// ServeHTTP 就是从树里面找节点
// 找到了就执行
func (h *HandlerBasedOnTree) serve(c *Context) {
	handler, found := h.search(c.Request.Method, c.Request.URL.Path)
	if !found {
		c.Writer.WriteHeader(http.StatusNotFound)
		_, _ = c.Writer.Write([]byte("404 Not Found"))
		return
	}
	handler(c)
}

func (h *HandlerBasedOnTree) search(method, path string) (handleFunc, bool) {
	root, ok := h.methodTrees[method]
	if !ok {
		return nil, false // 根路由都没有
	}

	if path == "/" {
		if root.handler != nil {
			return root.handler, true
		} else {
			return nil, false // 根路由没有 handleFunc
		}
	}

	// 去除头尾可能有的/，然后按照/切割成段
	paths := strings.Split(strings.Trim(path, "/"), "/")
	for _, p := range paths {
		// 从子节点里边找一个匹配到了当前 path 的节点
		matchChild, found := h.matchChild(root, p)
		if !found {
			return nil, false
		}
		root = matchChild
	}
	// 到这里，应该是找完了
	if root.handler == nil {
		// 到达这里是因为这种场景
		// 比如说你注册了 /user/profile
		// 然后你访问 /user
		return nil, false
	}

	return root.handler, true
}

// Route 就相当于往树里面插入节点
func (h *HandlerBasedOnTree) addRoute(method string, path string,
	handleFunc handleFunc) {
	// 校验
	h.validatePattern(path)

	root, ok := h.methodTrees[method]
	if !ok {
		root = &node{path: "/"}
		h.methodTrees[method] = root
	}

	if path == "/" {
		if root.handler != nil {
			panic("根路由冲突! ")
		}
		root.handler = handleFunc
	}

	// 将path 按照 URL 的分隔符切割
	// 例如，/user/friends 将变成 [user, friends]
	segs := strings.Split(path[1:], "/")
	// 开始一段段处理
	for _, s := range segs {
		if s == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		//filters = filters.childOrCreate(s)
		// 从子节点里边找一个匹配到了当前 path 的节点
		matchChild, found := h.matchChild(root, s)
		if found {
			root = matchChild
		} else {
			// 为当前节点根据
			root = h.buildSubTree(root, s, handleFunc)
		}
	}

	if root.handler != nil {
		panic("路由冲突! ")
	}

	root.handler = handleFunc
}

func (h *HandlerBasedOnTree) validatePattern(path string) {
	if path == "" {
		panic("web: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	// 校验 *，如果存在，必须在最后一个，并且它前面必须是/
	// 即我们只接受 /* 的存在，abc*这种是非法
	pos := strings.Index(path, "*")
	// 找到了 *
	if pos > 0 {
		// 确保只能是 /*
		if pos != len(path)-1 || path[pos-1] != '/' {
			panic("无效路由 ! ")
		}
	}

}

func (h *HandlerBasedOnTree) matchChild(root *node, path string) (*node, bool) {
	var wildcardNode *node
	for _, child := range root.children {
		// 并不是 * 的节点命中了，直接返回
		// != * 是为了防止用户乱输入
		if child.path == path &&
			child.path != "*" {
			return child, true
		}
		// 命中了通配符的，我们看看后面还有没有更加详细的
		// 比如访问 /user/profile, /user/* 在前面, /user/profile 在后面, 就无法正确匹配
		// 所以先不 return
		if child.path == "*" {
			wildcardNode = child
		}
	}
	return wildcardNode, wildcardNode != nil
}

func (h *HandlerBasedOnTree) buildSubTree(root *node, path string, handlerFn handleFunc) *node {
	cur := root
	nn := newNode(path)
	nn.handler = handlerFn

	cur.children = append(cur.children, nn)
	return nn
}
