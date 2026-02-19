package trie

import (
	"strings"

	"github.com/elmq0022/gohan/set"
)

type Node struct {
	ch   map[string]*Node
	subs *set.Set[int64]
}

func NewNode() *Node {
	return &Node{
		ch: make(map[string]*Node),
	}
}

func (n *Node) AddSub(sub string, sid int64) error {
	parts, err := validSub(sub)
	if err != nil {
		return err
	}
	cur := n
	for _, part := range parts {
		if cur.ch[part] == nil {
			cur.ch[part] = NewNode()
		}
		cur = cur.ch[part]
	}
	if cur.subs == nil {
		cur.subs = set.NewSet[int64]()
	}
	cur.subs.Add(sid)
	return nil
}

func (n *Node) Lookup(sub string) []int64 {
	res := set.NewSet[int64]()
	parts := strings.Split(sub, ".")
	match(parts, n, res)
	return res.Slice()
}

func match(parts []string, n *Node, res *set.Set[int64]) {
	if len(parts) == 0 {
		if n.subs != nil {
			res.Merge(n.subs)
		}
		return
	}

	if n.ch[parts[0]] != nil {
		match(parts[1:], n.ch[parts[0]], res)
	}

	if n.ch["*"] != nil {
		match(parts[1:], n.ch["*"], res)
	}

	if n.ch[">"] != nil && n.ch[">"].subs != nil {
		res.Merge(n.ch[">"].subs)
	}
}
