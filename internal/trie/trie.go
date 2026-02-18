package trie

import "strings"

type Node struct {
	ch   map[string]*Node
	subs []int
}

func NewNode() *Node {
	return &Node{
		ch: make(map[string]*Node),
	}
}

func (n *Node) AddSub(sub string, sid int) {
	parts := strings.Split(sub, ".")
	cur := n
	for _, part := range parts {
		if cur.ch[part] == nil {
			cur.ch[part] = NewNode()
		}
		cur = cur.ch[part]
	}
	cur.subs = append(cur.subs, sid)
}

func (n *Node) Lookup(sub string) []int {
	res := make([]int, 0)
	parts := strings.Split(sub, ".")
	match(parts, n, &res)
	return res
}

func match(parts []string, n *Node, res *[]int) {
	if len(parts) == 0 {
		*res = append(*res, n.subs...)
		return
	}

	if n.ch[parts[0]] != nil {
		match(parts[1:], n.ch[parts[0]], res)
	}

	if n.ch["*"] != nil {
		match(parts[1:], n.ch["*"], res)
	}

	if n.ch[">"] != nil {
		*res = append(*res, n.ch[">"].subs...)
	}
}
