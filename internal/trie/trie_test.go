package trie_test

import (
	"sort"
	"testing"

	"github.com/elmq0022/pub-sub/internal/trie"
)

func makeSub(sid int64) trie.Sub {
	return trie.Sub{SID: sid}
}

func sorted(subs []trie.Sub) []trie.Sub {
	sort.Slice(subs, func(i, j int) bool { return subs[i].SID < subs[j].SID })
	return subs
}

func mustLookup(t *testing.T, tr *trie.Trie, sub string) []trie.Sub {
	t.Helper()
	subs, err := tr.Lookup(sub)
	if err != nil {
		t.Fatalf("Lookup(%q) unexpected error: %v", sub, err)
	}
	return subs
}

func TestExactMatch(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.bar", makeSub(1))
	tr.AddSub("foo.bar", makeSub(2))

	got := sorted(mustLookup(t, tr, "foo.bar"))
	if len(got) != 2 || got[0].SID != 1 || got[1].SID != 2 {
		t.Fatalf("got %v, want SIDs [1 2]", got)
	}
}

func TestNoMatch(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.bar", makeSub(1))

	got := mustLookup(t, tr, "foo.baz")
	if len(got) != 0 {
		t.Fatalf("expected no matches, got %v", got)
	}
}

func TestWildcardStar(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.*", makeSub(10))

	got := mustLookup(t, tr, "foo.bar")
	if len(got) != 1 || got[0].SID != 10 {
		t.Fatalf("got %v, want [10]", got)
	}

	got = mustLookup(t, tr, "foo.baz")
	if len(got) != 1 || got[0].SID != 10 {
		t.Fatalf("got %v, want [10]", got)
	}

	got = mustLookup(t, tr, "foo.bar.baz")
	if len(got) != 0 {
		t.Fatalf("* should not match multiple levels, got %v", got)
	}
}

func TestWildcardGreaterThan(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.>", makeSub(20))

	for _, topic := range []string{"foo.bar", "foo.bar.baz", "foo.a.b.c"} {
		got := mustLookup(t, tr, topic)
		if len(got) != 1 || got[0].SID != 20 {
			t.Fatalf("topic %q: got %v, want [20]", topic, got)
		}
	}

	got := mustLookup(t, tr, "foo")
	if len(got) != 0 {
		t.Fatalf("got %v, want no match for 'foo'", got)
	}
}

func TestSameSubTwoPatterns(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.bar", makeSub(5))
	tr.AddSub("foo.*", makeSub(5))

	// trie returns all matches; dedup is the caller's responsibility
	got := mustLookup(t, tr, "foo.bar")
	if len(got) != 2 {
		t.Fatalf("expected 2 matches (no trie-level dedup), got %v", got)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("a.b", makeSub(1))
	tr.AddSub("a.*", makeSub(2))
	tr.AddSub("a.>", makeSub(3))

	got := sorted(mustLookup(t, tr, "a.b"))
	wantSIDs := []int64{1, 2, 3}
	if len(got) != len(wantSIDs) {
		t.Fatalf("got %v, want SIDs %v", got, wantSIDs)
	}
	for i, sid := range wantSIDs {
		if got[i].SID != sid {
			t.Fatalf("got %v, want SIDs %v", got, wantSIDs)
		}
	}
}

func TestRootGreaterThan(t *testing.T) {
	topics := []string{"foo", "foo.bar", "foo.bar.baz", "a.b.c.d"}

	tr := trie.NewTrie()
	tr.AddSub(">", makeSub(99))

	for _, topic := range topics {
		got := mustLookup(t, tr, topic)
		if len(got) != 1 || got[0].SID != 99 {
			t.Fatalf("topic %q: got %v, want [99]", topic, got)
		}
	}

	// Without > subscribed the same topics must return no matches.
	empty := trie.NewTrie()
	for _, topic := range topics {
		got := mustLookup(t, empty, topic)
		if len(got) != 0 {
			t.Fatalf("empty trie, topic %q: got %v, want []", topic, got)
		}
	}
}

func TestLookupInvalidSub(t *testing.T) {
	tr := trie.NewTrie()
	_, err := tr.Lookup("")
	if err == nil {
		t.Fatal("expected error for empty subject, got nil")
	}
}

// helpers

func makeSubFull(cid, sid int64) trie.Sub {
	return trie.Sub{CID: cid, SID: sid}
}

func mustAddSub(t *testing.T, tr *trie.Trie, pattern string, s trie.Sub) {
	t.Helper()
	if err := tr.AddSub(pattern, s); err != nil {
		t.Fatalf("AddSub(%q, %v) unexpected error: %v", pattern, s, err)
	}
}

func mustRemoveSub(t *testing.T, tr *trie.Trie, cid, sid int64) {
	t.Helper()
	if err := tr.RemoveSub(cid, sid); err != nil {
		t.Fatalf("RemoveSub(%d, %d) unexpected error: %v", cid, sid, err)
	}
}

// --- RemoveSub error cases ---

func TestRemoveSub_UnknownCID(t *testing.T) {
	tr := trie.NewTrie()
	err := tr.RemoveSub(99, 1)
	if err == nil {
		t.Fatal("expected error for unknown CID, got nil")
	}
}

func TestRemoveSub_UnknownSID(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "foo.bar", makeSubFull(1, 1))
	err := tr.RemoveSub(1, 99)
	if err == nil {
		t.Fatal("expected error for unknown SID, got nil")
	}
}

func TestRemoveSub_SecondRemoveFails(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "foo.bar", makeSubFull(1, 1))
	mustRemoveSub(t, tr, 1, 1)
	if err := tr.RemoveSub(1, 1); err == nil {
		t.Fatal("expected error on second RemoveSub, got nil")
	}
}

// --- basic removal ---

func TestRemoveSub_SubNoLongerMatches(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "foo.bar", makeSubFull(1, 1))
	mustRemoveSub(t, tr, 1, 1)

	got := mustLookup(t, tr, "foo.bar")
	if len(got) != 0 {
		t.Fatalf("expected no matches after removal, got %v", got)
	}
}

func TestRemoveSub_OtherSubsAtSameNodeUnaffected(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "foo.bar", makeSubFull(1, 1))
	mustAddSub(t, tr, "foo.bar", makeSubFull(2, 2))
	mustRemoveSub(t, tr, 1, 1)

	got := mustLookup(t, tr, "foo.bar")
	if len(got) != 1 || got[0].SID != 2 {
		t.Fatalf("expected only SID 2 after removal, got %v", got)
	}
}

func TestRemoveSub_OneSubOfClientLeavesOtherIntact(t *testing.T) {
	tr := trie.NewTrie()
	// same client (CID=1), two different subscriptions
	mustAddSub(t, tr, "foo.bar", makeSubFull(1, 1))
	mustAddSub(t, tr, "foo.baz", makeSubFull(1, 2))
	mustRemoveSub(t, tr, 1, 1)

	if got := mustLookup(t, tr, "foo.bar"); len(got) != 0 {
		t.Fatalf("foo.bar: expected no matches, got %v", got)
	}
	got := mustLookup(t, tr, "foo.baz")
	if len(got) != 1 || got[0].SID != 2 {
		t.Fatalf("foo.baz: expected SID 2, got %v", got)
	}
}

// --- index management ---

func TestRemoveSub_LastSIDForCIDAllowsOtherClients(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "a.b", makeSubFull(1, 10))
	mustAddSub(t, tr, "a.b", makeSubFull(2, 20))

	// remove the only sub for CID 1
	mustRemoveSub(t, tr, 1, 10)

	// CID 2's sub must still be reachable
	got := mustLookup(t, tr, "a.b")
	if len(got) != 1 || got[0].SID != 20 {
		t.Fatalf("expected SID 20 for CID 2 after CID 1 removed, got %v", got)
	}

	// CID 2's sub must still be removable via index
	if err := tr.RemoveSub(2, 20); err != nil {
		t.Fatalf("expected SID 20 to still be indexed, got error: %v", err)
	}
}

func TestRemoveSub_RemainingSIDsForSameCIDStillIndexed(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "a.b", makeSubFull(1, 1))
	mustAddSub(t, tr, "a.c", makeSubFull(1, 2))

	mustRemoveSub(t, tr, 1, 1)

	// SID 2 for CID 1 must still be indexed and removable
	if err := tr.RemoveSub(1, 2); err != nil {
		t.Fatalf("SID 2 should still be in index after SID 1 removed: %v", err)
	}
}

// --- node pruning ---

func TestRemoveSub_PrunesLeafNode(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "a.b.c", makeSubFull(1, 1))
	mustRemoveSub(t, tr, 1, 1)

	// Re-add a different sub; it should be the only result (no ghost subs).
	mustAddSub(t, tr, "a.b.c", makeSubFull(2, 2))
	got := mustLookup(t, tr, "a.b.c")
	if len(got) != 1 || got[0].SID != 2 {
		t.Fatalf("expected exactly 1 sub after prune+re-add, got %v", got)
	}
}

func TestRemoveSub_PrunesChainOfEmptyNodes(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "a.b.c.d", makeSubFull(1, 1))
	mustRemoveSub(t, tr, 1, 1)

	if got := mustLookup(t, tr, "a.b.c.d"); len(got) != 0 {
		t.Fatalf("expected empty after full-chain prune, got %v", got)
	}

	// An intermediate node can now be used cleanly.
	mustAddSub(t, tr, "a.b", makeSubFull(2, 2))
	got := mustLookup(t, tr, "a.b")
	if len(got) != 1 || got[0].SID != 2 {
		t.Fatalf("expected SID 2 at a.b after chain prune, got %v", got)
	}
}

func TestRemoveSub_NoPruneWhenNodeHasChildren(t *testing.T) {
	tr := trie.NewTrie()
	// a.b has a sub AND a child (a.b.c)
	mustAddSub(t, tr, "a.b", makeSubFull(1, 1))
	mustAddSub(t, tr, "a.b.c", makeSubFull(2, 2))

	mustRemoveSub(t, tr, 1, 1)

	// a.b.c must still be reachable (a.b node was NOT pruned)
	got := mustLookup(t, tr, "a.b.c")
	if len(got) != 1 || got[0].SID != 2 {
		t.Fatalf("expected SID 2 at a.b.c after removing a.b sub, got %v", got)
	}
}

func TestRemoveSub_NoPruneWhenOtherSubsRemain(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "a.b", makeSubFull(1, 1))
	mustAddSub(t, tr, "a.b", makeSubFull(2, 2))

	mustRemoveSub(t, tr, 1, 1)

	// a.b still has SID 2 - must not be pruned
	got := mustLookup(t, tr, "a.b")
	if len(got) != 1 || got[0].SID != 2 {
		t.Fatalf("expected SID 2 at a.b after partial removal, got %v", got)
	}
}

// --- wildcard subjects ---

func TestRemoveSub_WildcardStar(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "foo.*", makeSubFull(1, 1))
	mustAddSub(t, tr, "foo.*", makeSubFull(2, 2))
	mustRemoveSub(t, tr, 1, 1)

	got := mustLookup(t, tr, "foo.bar")
	if len(got) != 1 || got[0].SID != 2 {
		t.Fatalf("expected only SID 2 after removing * sub, got %v", got)
	}
}

func TestRemoveSub_WildcardStar_NoMatchAfterOnlySubRemoved(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "foo.*", makeSubFull(1, 1))
	mustRemoveSub(t, tr, 1, 1)

	got := mustLookup(t, tr, "foo.bar")
	if len(got) != 0 {
		t.Fatalf("expected no matches after sole * sub removed, got %v", got)
	}
}

func TestRemoveSub_WildcardGreaterThan(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "foo.>", makeSubFull(1, 1))
	mustAddSub(t, tr, "foo.>", makeSubFull(2, 2))
	mustRemoveSub(t, tr, 1, 1)

	for _, topic := range []string{"foo.bar", "foo.bar.baz"} {
		got := mustLookup(t, tr, topic)
		if len(got) != 1 || got[0].SID != 2 {
			t.Fatalf("topic %q: expected only SID 2 after removing > sub, got %v", topic, got)
		}
	}
}

func TestRemoveSub_WildcardGreaterThan_NoMatchAfterOnlySubRemoved(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "foo.>", makeSubFull(1, 1))
	mustRemoveSub(t, tr, 1, 1)

	for _, topic := range []string{"foo.bar", "foo.bar.baz"} {
		got := mustLookup(t, tr, topic)
		if len(got) != 0 {
			t.Fatalf("topic %q: expected no matches after sole > sub removed, got %v", topic, got)
		}
	}
}

// --- re-add after remove ---

func TestRemoveSub_ReAddAfterRemoveYieldsOneResult(t *testing.T) {
	tr := trie.NewTrie()
	mustAddSub(t, tr, "x.y", makeSubFull(1, 1))
	mustRemoveSub(t, tr, 1, 1)
	mustAddSub(t, tr, "x.y", makeSubFull(1, 1))

	got := mustLookup(t, tr, "x.y")
	if len(got) != 1 || got[0].SID != 1 {
		t.Fatalf("expected exactly 1 result after re-add, got %v", got)
	}
}
