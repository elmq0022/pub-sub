package trie

import (
	"errors"
	"sync"
)

type client struct{}

type Sub struct {
	CID    int64
	SID    int64
	Client *client
}

type node struct {
	key      string
	parent   *node
	children map[string]*node
	subs     []Sub
}

func newNode(parent *node, key string) *node {
	return &node{
		key:      key,
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
		root:  newNode(nil, ""),
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
			cur.children[part] = newNode(cur, part)
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

func (t *Trie) RemoveSub(CID, SID int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.index[CID] == nil || t.index[CID][SID] == nil {
		return errors.New("no subscription")
	}

	n := t.index[CID][SID]
	delete(t.index[CID], SID)
	if len(t.index[CID]) == 0 {
		delete(t.index, CID)
	}

	n.removeSub(CID, SID)

	for n != nil && n.parent != nil && len(n.subs) == 0 && len(n.children) == 0 {
		parent := n.parent
		delete(parent.children, n.key)
		n = parent
	}

	return nil
}

// removeSub removes all subs matching (CID, SID) from the node's subs slice in
// O(n) time. Duplicates are possible because AddSub does not check for them,
// so this cleans up all copies in a single pass.
func (n *node) removeSub(CID, SID int64) {
	if n == nil || len(n.subs) == 0 {
		return
	}

	l, r := 0, len(n.subs)-1
	for l <= r {
		for r >= l && n.subs[r].CID == CID && n.subs[r].SID == SID {
			r--
		}
		if l < r && n.subs[l].CID == CID && n.subs[l].SID == SID {
			n.subs[l], n.subs[r] = n.subs[r], n.subs[l]
		}
		l++
	}

	n.subs = n.subs[:r+1]
}
