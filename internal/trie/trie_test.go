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

func TestExactMatch(t *testing.T) {
	root := trie.NewNode()
	root.AddSub("foo.bar", 1)
	root.AddSub("foo.bar", 2)

	got := sorted(root.Lookup("foo.bar"))
	want := []int64{1, 2}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestNoMatch(t *testing.T) {
	root := trie.NewNode()
	root.AddSub("foo.bar", 1)

	got := root.Lookup("foo.baz")
	if len(got) != 0 {
		t.Fatalf("expected no matches, got %v", got)
	}
}

func TestWildcardStar(t *testing.T) {
	root := trie.NewNode()
	root.AddSub("foo.*", 10)

	got := root.Lookup("foo.bar")
	if len(got) != 1 || got[0] != 10 {
		t.Fatalf("got %v, want [10]", got)
	}

	got = root.Lookup("foo.baz")
	if len(got) != 1 || got[0] != 10 {
		t.Fatalf("got %v, want [10]", got)
	}

	got = root.Lookup("foo.bar.baz")
	if len(got) != 0 {
		t.Fatalf("* should not match multiple levels, got %v", got)
	}
}

func TestWildcardGreaterThan(t *testing.T) {
	root := trie.NewNode()
	root.AddSub("foo.>", 20)

	for _, topic := range []string{"foo.bar", "foo.bar.baz", "foo.a.b.c"} {
		got := root.Lookup(topic)
		if len(got) != 1 || got[0] != 20 {
			t.Fatalf("topic %q: got %v, want [20]", topic, got)
		}
	}

	got := root.Lookup("foo")
	if len(got) != 0 {
		t.Fatalf("got %v, want no match for 'foo'", got)
	}
}

func TestDeduplicate(t *testing.T) {
	root := trie.NewNode()
	root.AddSub("foo.bar", 5)
	root.AddSub("foo.*", 5)

	got := root.Lookup("foo.bar")
	if len(got) != 1 || got[0] != 5 {
		t.Fatalf("expected deduplication, got %v", got)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	root := trie.NewNode()
	root.AddSub("a.b", 1)
	root.AddSub("a.*", 2)
	root.AddSub("a.>", 3)

	got := sorted(root.Lookup("a.b"))
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
