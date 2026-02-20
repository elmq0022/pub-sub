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
