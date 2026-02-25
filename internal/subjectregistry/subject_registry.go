package subjectregistry

import (
	"errors"
	"strings"
)

type Sub struct {
	CID int64
	SID int64
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

type SubjectRegistry struct {
	root  *node
	index map[int64]map[int64]*node
}

type Registry interface {
	AddSub(subject string, s Sub) error
	Lookup(subject string) ([]Sub, error)
	RemoveSub(CID, SID int64) error
	RemoveCID(CID int64) error
}

func NewSubjectRegistry() *SubjectRegistry {
	return &SubjectRegistry{
		root:  newNode(nil, ""),
		index: make(map[int64]map[int64]*node),
	}
}

func (t *SubjectRegistry) AddSub(subject string, s Sub) error {
	cur := t.root

	parts := strings.Split(subject, ".")
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

func (t *SubjectRegistry) Lookup(subject string) ([]Sub, error) {
	parts := strings.Split(subject, ".")

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

func (t *SubjectRegistry) RemoveSub(CID, SID int64) error {
	if t.index[CID] == nil || t.index[CID][SID] == nil {
		return errors.New("no subscription")
	}

	n := t.index[CID][SID]
	delete(t.index[CID], SID)
	if len(t.index[CID]) == 0 {
		delete(t.index, CID)
	}

	t.removeSubFromNodeAndPrune(n, CID, SID)

	return nil
}

func (t *SubjectRegistry) RemoveCID(CID int64) error {
	bySID := t.index[CID]
	if bySID == nil {
		return nil
	}

	for sid, n := range bySID {
		t.removeSubFromNodeAndPrune(n, CID, sid)
	}
	delete(t.index, CID)

	return nil
}

func (t *SubjectRegistry) removeSubFromNodeAndPrune(n *node, CID, SID int64) {
	n.removeSub(CID, SID)

	for n != nil && n.parent != nil && len(n.subs) == 0 && len(n.children) == 0 {
		parent := n.parent
		delete(parent.children, n.key)
		n = parent
	}
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
