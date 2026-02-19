package trie_test

import (
	"sort"
	"testing"

	"github.com/elmq0022/pub-sub/internal/trie"
)

func sorted(ids []int64) []int64 {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func mustLookup(t *testing.T, tr *trie.Trie, sub string) []int64 {
	t.Helper()
	ids, err := tr.Lookup(sub)
	if err != nil {
		t.Fatalf("Lookup(%q) unexpected error: %v", sub, err)
	}
	return ids
}

func TestExactMatch(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.bar", 1)
	tr.AddSub("foo.bar", 2)

	got := sorted(mustLookup(t, tr, "foo.bar"))
	want := []int64{1, 2}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestNoMatch(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.bar", 1)

	got := mustLookup(t, tr, "foo.baz")
	if len(got) != 0 {
		t.Fatalf("expected no matches, got %v", got)
	}
}

func TestWildcardStar(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.*", 10)

	got := mustLookup(t, tr, "foo.bar")
	if len(got) != 1 || got[0] != 10 {
		t.Fatalf("got %v, want [10]", got)
	}

	got = mustLookup(t, tr, "foo.baz")
	if len(got) != 1 || got[0] != 10 {
		t.Fatalf("got %v, want [10]", got)
	}

	got = mustLookup(t, tr, "foo.bar.baz")
	if len(got) != 0 {
		t.Fatalf("* should not match multiple levels, got %v", got)
	}
}

func TestWildcardGreaterThan(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.>", 20)

	for _, topic := range []string{"foo.bar", "foo.bar.baz", "foo.a.b.c"} {
		got := mustLookup(t, tr, topic)
		if len(got) != 1 || got[0] != 20 {
			t.Fatalf("topic %q: got %v, want [20]", topic, got)
		}
	}

	got := mustLookup(t, tr, "foo")
	if len(got) != 0 {
		t.Fatalf("got %v, want no match for 'foo'", got)
	}
}

func TestDeduplicate(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("foo.bar", 5)
	tr.AddSub("foo.*", 5)

	got := mustLookup(t, tr, "foo.bar")
	if len(got) != 1 || got[0] != 5 {
		t.Fatalf("expected deduplication, got %v", got)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	tr := trie.NewTrie()
	tr.AddSub("a.b", 1)
	tr.AddSub("a.*", 2)
	tr.AddSub("a.>", 3)

	got := sorted(mustLookup(t, tr, "a.b"))
	want := []int64{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
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
