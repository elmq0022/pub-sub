package codec

import "testing"

func walkTransitionTable(input string) (STATE, int) {
	state := ST_START
	for i := 0; i < len(input); i++ {
		state = transitionTable[state][input[i]]
		if state == ST_ERROR {
			return state, i
		}
	}
	return state, -1
}

func TestTransitionTableAcceptsImplementedProtocols(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantState STATE
	}{
		{name: "ping", input: "PING\r\n", wantState: ST_DONE},
		{name: "pong", input: "PONG\r\n", wantState: ST_DONE},
		{name: "connect empty json", input: "CONNECT {}\r\n", wantState: ST_DONE},
		{name: "sub simple", input: "SUB foo 1\r\n", wantState: ST_DONE},
		{name: "sub dotted", input: "SUB foo.bar 42\r\n", wantState: ST_DONE},
		{name: "sub star wildcard", input: "SUB foo.* 7\r\n", wantState: ST_DONE},
		{name: "sub gt wildcard", input: "SUB foo.> 7\r\n", wantState: ST_DONE},
		{name: "sub root gt wildcard", input: "SUB > 9\r\n", wantState: ST_DONE},
		{name: "pub header simple", input: "PUB foo 0\r\n", wantState: ST_PUB_PAYLOAD},
		{name: "pub header dotted", input: "PUB foo.bar 12\r\n", wantState: ST_PUB_PAYLOAD},
		{name: "unsub simple", input: "UNSUB 1\r\n", wantState: ST_DONE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, failIdx := walkTransitionTable(tt.input)
			if failIdx != -1 {
				t.Fatalf("input rejected at byte %d (%q)", failIdx, tt.input[failIdx])
			}
			if got != tt.wantState {
				t.Fatalf("final state = %v, want %v", got, tt.wantState)
			}
		})
	}
}

func TestTransitionTableRejectsUnsupportedOrMalformedProtocols(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "ping without cr", input: "PING\n"},
		{name: "connect without required space", input: "CONNECT{}\r\n"},
		{name: "connect with optional args unsupported", input: "CONNECT {\"verbose\":false}\r\n"},
		{name: "sub missing sid", input: "SUB foo\r\n"},
		{name: "sub leading dot", input: "SUB .foo 1\r\n"},
		{name: "sub empty token", input: "SUB foo..bar 1\r\n"},
		{name: "sub gt not terminal", input: "SUB foo.>.bar 1\r\n"},
		{name: "pub missing bytes", input: "PUB foo\r\n"},
		{name: "pub with optional reply-to unsupported", input: "PUB foo reply 5\r\n"},
		{name: "unsub missing sid", input: "UNSUB\r\n"},
		{name: "unsub with optional max msgs unsupported", input: "UNSUB 1 2\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, failIdx := walkTransitionTable(tt.input)
			if failIdx == -1 {
				t.Fatalf("expected rejection, but parser ended in state %v", got)
			}
		})
	}
}
