package trie

import "sync"

type client struct{}

type Sub struct {
	CID    int64
	SID    int64
	Client *client
}

type node struct {
	parent   *node
	children map[string]*node
	subs     []Sub
}

func newNode(parent *node) *node {
	return &node{
		parent:   parent,
		children: make(map[string]*node)}
}

type Trie struct {
	mu    sync.RWMutex
	root  *node
	index map[int64]map[int64]*node
}

func NewTrie() *Trie {
	return &Trie{
		root:  newNode(nil),
		index: make(map[int64]map[int64]*node),
	}
}

func (t *Trie) AddSub(sub string, s Sub) error {
	parts, err := validSub(sub)
	if err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.root
	for _, part := range parts {
		if cur.children[part] == nil {
			cur.children[part] = newNode(cur)
		}
		cur = cur.children[part]
	}
	if t.index[s.CID] == nil {
		t.index[s.CID] = make(map[int64]*node)
	}
	t.index[s.CID][s.SID] = cur
	cur.subs = append(cur.subs, s)
	return nil
}

func (t *Trie) Lookup(sub string) ([]Sub, error) {
	parts, err := validLookup(sub)
	if err != nil {
		return nil, err
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	var res []Sub
	match(parts, t.root, &res)
	return res, nil
}

func match(parts []string, n *node, res *[]Sub) {
	if len(parts) == 0 {
		*res = append(*res, n.subs...)
		return
	}

	if n.children[parts[0]] != nil {
		match(parts[1:], n.children[parts[0]], res)
	}

	if n.children["*"] != nil {
		match(parts[1:], n.children["*"], res)
	}

	if n.children[">"] != nil {
		*res = append(*res, n.children[">"].subs...)
	}
}
