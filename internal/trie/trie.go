package trie

import (
	"sync"

	"github.com/elmq0022/gohan/set"
)

type Node struct {
	ch   map[string]*Node
	subs *set.Set[int64]
}

func newNode() *Node {
	return &Node{
		ch: make(map[string]*Node),
	}
}

type Trie struct {
	mu   sync.RWMutex
	root *Node
}

func NewTrie() *Trie {
	return &Trie{root: newNode()}
}

func (t *Trie) AddSub(sub string, sid int64) error {
	parts, err := validSub(sub)
	if err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.root
	for _, part := range parts {
		if cur.ch[part] == nil {
			cur.ch[part] = newNode()
		}
		cur = cur.ch[part]
	}
	if cur.subs == nil {
		cur.subs = set.NewSet[int64]()
	}
	cur.subs.Add(sid)
	return nil
}

func (t *Trie) Lookup(sub string) ([]int64, error) {
	parts, err := validSub(sub)
	if err != nil {
		return nil, err
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	res := set.NewSet[int64]()
	match(parts, t.root, res)
	return res.Slice(), nil
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
