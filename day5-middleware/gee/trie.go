package gee

import "strings"

// trie 树节点
type node struct {
	pattern  string  // 待匹配路由, 例如 /p/:lang
	part     string  // 路由中的一部分, 例如 :lang
	children []*node // 子节点, 例如 [doc, tutorial, intro]
	isWild   bool    // 是否精确匹配, part 含有 : 或 * 时为true
}

// 第一个匹配成功的节点, 用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点, 用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)

	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}

	return nodes
}

// 插入节点(example: pattern: /p/:lang, parts: [p, :lang], height: 0)
func (n *node) insert(pattern string, parts []string, height int) {
	// 递归终止条件
	if len(parts) == height {
		// 已经存在, example: /p/:lang 和 /p/a
		if n.pattern != "" {
			panic("duplicate pattern: " + n.pattern + " " + pattern)
		}

		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)

	// 不存在则创建
	if child == nil {
		// 不要变成 :=, 否则会创建局部变量
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}

	// 递归
	child.insert(pattern, parts, height+1)
}

// 查找节点(example: parts: [p, :lang], height: 0)
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") { // 匹配到了或者遇到通配符
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}
