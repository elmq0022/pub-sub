package broker

import (
	"github.com/elmq0022/pub-sub/internal/codec"
	"github.com/elmq0022/pub-sub/internal/subjectregistry"
)

type BrokerEvent interface{ isBrokerEvent() }

// Wire command from reader actor.
type CmdEvent struct {
	CID int64
	Cmd codec.InboundCommands
}

func (CmdEvent) isBrokerEvent() {}

type ProtocolErrorEvent struct {
	CID int64
	Msg string
}

func (ProtocolErrorEvent) isBrokerEvent() {}

type SessionUpEvent struct {
	CID      int64
	Outbound chan<- codec.OutboundCommands
}

func (SessionUpEvent) isBrokerEvent() {}

type SessionDownEvent struct {
	CID int64
}

func (SessionDownEvent) isBrokerEvent() {}

type Broker struct {
	registry subjectregistry.Registry
	outbound map[int64]chan<- codec.OutboundCommands
	inbox    chan BrokerEvent
}

func NewBroker(r subjectregistry.Registry) *Broker {
	return &Broker{
		registry: r,
		outbound: make(map[int64]chan<- codec.OutboundCommands),
		inbox:    make(chan BrokerEvent),
	}
}

func (b *Broker) Input() chan<- BrokerEvent {
	return b.inbox
}

func (b *Broker) Run() {
	for msg := range b.inbox {
		switch ev := msg.(type) {
		case CmdEvent:
			b.handleCmdEvent(ev)
		case ProtocolErrorEvent:
			b.handleProtocolErrorEvent(ev)
		case SessionUpEvent:
			b.handleSessionUpEvent(ev)
		case SessionDownEvent:
			b.handleSessionDownEvent(ev)
		}
	}
}

func (b *Broker) handleSessionUpEvent(ev SessionUpEvent) {
	b.outbound[ev.CID] = ev.Outbound
}

func (b *Broker) handleSessionDownEvent(ev SessionDownEvent) {
	if outbox, ok := b.outbound[ev.CID]; ok {
		close(outbox)
		delete(b.outbound, ev.CID)
	}
	b.registry.RemoveCID(ev.CID)
}

func (b *Broker) disconnectCID(cid int64, outbox chan<- codec.OutboundCommands) {
	close(outbox)
	delete(b.outbound, cid)
	b.registry.RemoveCID(cid)
}

func (b *Broker) handleCmdEvent(ev CmdEvent) {
	switch cmd := ev.Cmd.(type) {
	case codec.Ping:
		outbox, ok := b.outbound[ev.CID]
		if !ok {
			break
		}
		select {
		case outbox <- codec.Pong{}:
		default:
			b.disconnectCID(ev.CID, outbox)
		}
	case codec.Pong:
		// TODO:
	case codec.Connect:
		// TODO:
	case codec.Sub:
		b.registry.AddSub(
			string(cmd.Subject),
			subjectregistry.Sub{
				CID: ev.CID,
				SID: cmd.SID,
			},
		)
	case codec.Pub:
		subs, err := b.registry.Lookup(string(cmd.Subject))
		if err != nil {
			break
		}
		for _, sub := range subs {
			outbox, ok := b.outbound[sub.CID]
			if !ok {
				continue
			}
			msg := codec.Msg{
				Subject: cmd.Subject,
				SID:     sub.SID,
				Payload: cmd.Payload,
			}
			select {
			case outbox <- msg:
			default:
				b.disconnectCID(sub.CID, outbox)
			}
		}

	case codec.Unsub:
		b.registry.RemoveSub(ev.CID, cmd.SID)
	}
}

func (b *Broker) handleProtocolErrorEvent(ev ProtocolErrorEvent) {
	outbox, ok := b.outbound[ev.CID]
	if !ok {
		b.registry.RemoveCID(ev.CID)
		return
	}

	select {
	case outbox <- codec.Err{Message: ev.Msg}:
	default:
	}

	b.disconnectCID(ev.CID, outbox)
}
