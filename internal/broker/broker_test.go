package broker

import (
	"testing"
	"time"

	"github.com/elmq0022/pub-sub/internal/codec"
	"github.com/elmq0022/pub-sub/internal/config"
	"github.com/elmq0022/pub-sub/internal/subjectregistry"
)

func TestHandleSessionDownEventRemovesSessionAndSubscriptions(t *testing.T) {
	registry := subjectregistry.NewSubjectRegistry()
	b := NewBroker(registry, testConfig())

	outbound := make(chan codec.OutboundCommands, 1)
	b.handleSessionUpEvent(SessionUpEvent{
		CID:      42,
		Outbound: outbound,
	})

	b.handleCmdEvent(CmdEvent{
		CID: 42,
		Cmd: codec.Sub{Subject: []byte("foo.bar"), SID: 7},
	})
	assertOutboundOK(t, outbound)

	b.handleSessionDownEvent(SessionDownEvent{CID: 42})

	if _, ok := b.sessions[42]; ok {
		t.Fatal("session still present after SessionDownEvent")
	}

	assertClosed(t, outbound)

	subs, err := registry.Lookup("foo.bar")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("expected subscriptions to be removed, got %d", len(subs))
	}
}

func TestHandleSessionDownEventDuplicateIsIdempotent(t *testing.T) {
	registry := subjectregistry.NewSubjectRegistry()
	b := NewBroker(registry, testConfig())

	outbound := make(chan codec.OutboundCommands, 1)
	b.handleSessionUpEvent(SessionUpEvent{
		CID:      7,
		Outbound: outbound,
	})

	b.handleCmdEvent(CmdEvent{
		CID: 7,
		Cmd: codec.Sub{Subject: []byte("foo.one"), SID: 1},
	})
	assertOutboundOK(t, outbound)
	b.handleCmdEvent(CmdEvent{
		CID: 7,
		Cmd: codec.Sub{Subject: []byte("foo.two"), SID: 2},
	})
	assertOutboundOK(t, outbound)

	b.handleSessionDownEvent(SessionDownEvent{CID: 7})
	b.handleSessionDownEvent(SessionDownEvent{CID: 7})

	if _, ok := b.sessions[7]; ok {
		t.Fatal("session still present after duplicate SessionDownEvent")
	}

	assertClosed(t, outbound)

	for _, subject := range []string{"foo.one", "foo.two"} {
		subs, err := registry.Lookup(subject)
		if err != nil {
			t.Fatalf("lookup failed for %q: %v", subject, err)
		}
		if len(subs) != 0 {
			t.Fatalf("expected subscriptions for %q to be removed, got %d", subject, len(subs))
		}
	}
}

func TestHandleProtocolErrorEventDisconnectsSessionAndRemovesSubscriptions(t *testing.T) {
	registry := subjectregistry.NewSubjectRegistry()
	b := NewBroker(registry, testConfig())

	outbound := make(chan codec.OutboundCommands, 1)
	b.handleSessionUpEvent(SessionUpEvent{
		CID:      9,
		Outbound: outbound,
	})

	b.handleCmdEvent(CmdEvent{
		CID: 9,
		Cmd: codec.Sub{Subject: []byte("foo.err"), SID: 3},
	})
	assertOutboundOK(t, outbound)

	b.handleProtocolErrorEvent(ProtocolErrorEvent{
		CID: 9,
		Msg: "unparsable command",
	})

	if _, ok := b.sessions[9]; ok {
		t.Fatal("session still present after ProtocolErrorEvent")
	}

	msg, ok := readOutbound(t, outbound)
	if !ok {
		t.Fatal("expected outbound error before channel close")
	}
	errMsg, ok := msg.(codec.Err)
	if !ok {
		t.Fatalf("expected codec.Err, got %T", msg)
	}
	if errMsg.Message != "unparsable command" {
		t.Fatalf("expected error message to be preserved, got %q", errMsg.Message)
	}

	assertClosed(t, outbound)

	subs, err := registry.Lookup("foo.err")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("expected subscriptions to be removed, got %d", len(subs))
	}
}

func TestHandleProtocolErrorEventWithoutSessionRemovesStaleSubscriptions(t *testing.T) {
	registry := subjectregistry.NewSubjectRegistry()
	b := NewBroker(registry, testConfig())

	if err := registry.AddSub("foo.stale", subjectregistry.Sub{CID: 55, SID: 8}); err != nil {
		t.Fatalf("failed to seed registry: %v", err)
	}

	b.handleProtocolErrorEvent(ProtocolErrorEvent{
		CID: 55,
		Msg: "unparsable command",
	})

	subs, err := registry.Lookup("foo.stale")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("expected stale subscriptions to be removed, got %d", len(subs))
	}
}

func TestHandleCmdEventUnsubUnknownSIDAckAndKeepsSession(t *testing.T) {
	registry := subjectregistry.NewSubjectRegistry()
	b := NewBroker(registry, testConfig())

	outbound := make(chan codec.OutboundCommands, 1)
	b.handleSessionUpEvent(SessionUpEvent{
		CID:      11,
		Outbound: outbound,
	})

	b.handleCmdEvent(CmdEvent{
		CID: 11,
		Cmd: codec.Unsub{SID: 999},
	})

	assertOutboundOK(t, outbound)

	if _, ok := b.sessions[11]; !ok {
		t.Fatal("session removed after unknown UNSUB")
	}

	select {
	case msg, ok := <-outbound:
		if !ok {
			t.Fatal("outbound channel unexpectedly closed")
		}
		t.Fatalf("unexpected outbound message after UNSUB ack: %T", msg)
	default:
	}
}

func readOutbound(t *testing.T, ch <-chan codec.OutboundCommands) (codec.OutboundCommands, bool) {
	t.Helper()

	select {
	case msg, ok := <-ch:
		return msg, ok
	default:
		t.Fatal("expected outbound message")
		return nil, false
	}
}

func assertOutboundOK(t *testing.T, ch <-chan codec.OutboundCommands) {
	t.Helper()

	msg, ok := readOutbound(t, ch)
	if !ok {
		t.Fatal("expected codec.OK before channel close")
	}
	if _, ok := msg.(codec.OK); !ok {
		t.Fatalf("expected codec.OK, got %T", msg)
	}
}

func assertClosed(t *testing.T, ch <-chan codec.OutboundCommands) {
	t.Helper()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected outbound channel to be closed")
		}
	default:
		t.Fatal("expected outbound channel to be closed")
	}
}

func testConfig() config.Config {
	return config.Config{
		Port:                  "8080",
		HeartbeatTickInterval: time.Second,
		HeartbeatTimeout:      3 * time.Second,
	}
}
