package trie

import (
	"sync"

	"github.com/elmq0022/gohan/set"
)

type node struct {
	ch   map[string]*node
	subs *set.Set[int64]
}

func newNode() *node {
	return &node{
		ch: make(map[string]*node),
	}
}

type Trie struct {
	mu   sync.RWMutex
	root *node
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
	parts, err := validLookup(sub)
	if err != nil {
		return nil, err
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	res := set.NewSet[int64]()
	match(parts, t.root, res)
	return res.Slice(), nil
}

func match(parts []string, n *node, res *set.Set[int64]) {
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
