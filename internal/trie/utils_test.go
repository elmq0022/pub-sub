package trie

import (
	"errors"
	"testing"
)

func TestValidSub(t *testing.T) {
	tests := []struct {
		name    string
		sub     string
		want    []string
		wantErr string
	}{
		{
			name: "simple topic",
			sub:  "foo.bar.baz",
			want: []string{"foo", "bar", "baz"},
		},
		{
			name: "single token",
			sub:  "foo",
			want: []string{"foo"},
		},
		{
			name: "star wildcard",
			sub:  "foo.*",
			want: []string{"foo", "*"},
		},
		{
			name: "greater-than wildcard at end",
			sub:  "foo.>",
			want: []string{"foo", ">"},
		},
		{
			name: "greater-than only",
			sub:  ">",
			want: []string{">"},
		},
		{
			name:    "empty string",
			sub:     "",
			wantErr: "must not be empty",
		},
		{
			name:    "empty token (leading dot)",
			sub:     ".foo",
			wantErr: "empty token",
		},
		{
			name:    "empty token (trailing dot)",
			sub:     "foo.",
			wantErr: "empty token",
		},
		{
			name:    "empty token (double dot)",
			sub:     "foo..bar",
			wantErr: "empty token",
		},
		{
			name:    "greater-than not last",
			sub:     "foo.>.bar",
			wantErr: "'>' must be the last token",
		},
		{
			name:    "mixed wildcard in token",
			sub:     "foo.bar>",
			wantErr: "wildcards must be standalone tokens",
		},
		{
			name:    "star embedded in token",
			sub:     "foo.b*r",
			wantErr: "wildcards must be standalone tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validSub(tt.sub)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				var subErr *SubError
				if !errors.As(err, &subErr) {
					t.Fatalf("expected *SubError, got %T: %v", err, err)
				}
				if subErr.Reason != tt.wantErr {
					t.Fatalf("expected reason %q, got %q", tt.wantErr, subErr.Reason)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSubError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *SubError
		want string
	}{
		{
			name: "with sub",
			err:  &SubError{Sub: "foo.>", Reason: "'>' must be the last token"},
			want: `invalid sub "foo.>": '>' must be the last token`,
		},
		{
			name: "empty sub",
			err:  &SubError{Sub: "", Reason: "must not be empty"},
			want: "invalid sub: must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
