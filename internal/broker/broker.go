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
	inbox    <-chan BrokerEvent
}

func NewBroker(r subjectregistry.Registry) *Broker {
	return &Broker{
		registry: r,
		outbound: make(map[int64]chan<- codec.OutboundCommands),
		inbox:    make(<-chan BrokerEvent),
	}
}

func (b *Broker) Run() {
	for msg := range b.inbox {
		switch ev := msg.(type) {
		case CmdEvent:
			b.handleCmdEvent(ev)
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
	delete(b.outbound, ev.CID)
	b.registry.RemoveCID(ev.CID)
}

func (b *Broker) handleCmdEvent(ev CmdEvent) {
	switch cmd := ev.Cmd.(type) {
	case codec.Ping:
		outbox, ok := b.outbound[ev.CID]
		if !ok {
			break
		}
		outbox <- codec.Pong{}
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
			outbox <- codec.Msg{
				Subject: cmd.Subject,
				SID:     sub.SID,
				Payload: cmd.Payload,
			}
		}

	case codec.Unsub:
		b.registry.RemoveSub(ev.CID, cmd.SID)
	}
}
