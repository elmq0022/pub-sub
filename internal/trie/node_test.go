package trie

import (
	"testing"
)

func sub(cid, sid int64) Sub {
	return Sub{CID: cid, SID: sid}
}

func TestRemoveSub_NilNode(t *testing.T) {
	var n *node
	// should not panic
	n.removeSub(1, 1)
}

func TestRemoveSub_EmptySlice(t *testing.T) {
	n := newNode(nil, "")
	// subs is nil/empty - should not panic and remain empty
	n.removeSub(1, 1)
	if len(n.subs) != 0 {
		t.Fatalf("expected 0 subs, got %v", n.subs)
	}
}

func TestRemoveSub_OneElement_NoMatch(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(2, 2)}

	n.removeSub(1, 1)

	if len(n.subs) != 1 || n.subs[0].SID != 2 {
		t.Fatalf("expected sub {2,2} to remain, got %v", n.subs)
	}
}

func TestRemoveSub_OneElement_Match(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(1, 1)}

	n.removeSub(1, 1)

	if len(n.subs) != 0 {
		t.Fatalf("expected empty subs after removal, got %v", n.subs)
	}
}

func TestRemoveSub_ManyElements_AllMatch(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(1, 1), sub(1, 1), sub(1, 1)}

	n.removeSub(1, 1)

	if len(n.subs) != 0 {
		t.Fatalf("expected all subs removed, got %v", n.subs)
	}
}

func TestRemoveSub_ManyElements_NoneMatch(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(2, 2), sub(3, 3), sub(4, 4)}

	n.removeSub(1, 1)

	if len(n.subs) != 3 {
		t.Fatalf("expected 3 subs to remain, got %v", n.subs)
	}
}

func TestRemoveSub_ManyElements_SomeMatch_MatchAtStart(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(1, 1), sub(1, 1), sub(2, 2), sub(3, 3)}

	n.removeSub(1, 1)

	if len(n.subs) != 2 {
		t.Fatalf("expected 2 subs, got %v", n.subs)
	}
	for _, s := range n.subs {
		if s.CID == 1 && s.SID == 1 {
			t.Fatalf("removed sub still present in %v", n.subs)
		}
	}
}

func TestRemoveSub_ManyElements_SomeMatch_MatchAtEnd(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(2, 2), sub(3, 3), sub(1, 1), sub(1, 1)}

	n.removeSub(1, 1)

	if len(n.subs) != 2 {
		t.Fatalf("expected 2 subs, got %v", n.subs)
	}
	for _, s := range n.subs {
		if s.CID == 1 && s.SID == 1 {
			t.Fatalf("removed sub still present in %v", n.subs)
		}
	}
}

func TestRemoveSub_ManyElements_SomeMatch_MatchInterleaved(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(2, 2), sub(1, 1), sub(3, 3), sub(1, 1), sub(4, 4)}

	n.removeSub(1, 1)

	if len(n.subs) != 3 {
		t.Fatalf("expected 3 subs, got %v", n.subs)
	}
	for _, s := range n.subs {
		if s.CID == 1 && s.SID == 1 {
			t.Fatalf("removed sub still present in %v", n.subs)
		}
	}
}

// Only CID matches but SID differs - should not remove.
func TestRemoveSub_PartialKeyMatch_CIDOnly(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(1, 99)}

	n.removeSub(1, 1)

	if len(n.subs) != 1 || n.subs[0].SID != 99 {
		t.Fatalf("expected sub {1,99} to remain, got %v", n.subs)
	}
}

// Only SID matches but CID differs - should not remove.
func TestRemoveSub_PartialKeyMatch_SIDOnly(t *testing.T) {
	n := newNode(nil, "")
	n.subs = []Sub{sub(99, 1)}

	n.removeSub(1, 1)

	if len(n.subs) != 1 || n.subs[0].CID != 99 {
		t.Fatalf("expected sub {99,1} to remain, got %v", n.subs)
	}
}
