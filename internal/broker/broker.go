package broker

import (
	"time"

	"github.com/elmq0022/pub-sub/internal/codec"
	"github.com/elmq0022/pub-sub/internal/subjectregistry"
)

type ClientSession struct {
	Outbound     chan<- codec.OutboundCommands
	AwaitingPong bool
	PingSentAt   time.Time
}

type BrokerConfig struct {
	HeartbeatTickInterval time.Duration
	HeartbeatTimeout      time.Duration
}

type Broker struct {
	registry subjectregistry.Registry
	sessions map[int64]ClientSession
	inbox    chan BrokerEvent
	config   BrokerConfig
}

func NewBroker(r subjectregistry.Registry, config BrokerConfig) *Broker {
	return &Broker{
		registry: r,
		sessions: make(map[int64]ClientSession),
		inbox:    make(chan BrokerEvent),
		config:   config,
	}
}

func (b *Broker) Input() chan<- BrokerEvent {
	return b.inbox
}

func (b *Broker) Run() {
	go b.startHeartbeat()

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
		case HeartbeatTickEvent:
			b.handleHeartbeatTickEvent(ev)
		}
	}
}

func (b *Broker) startHeartbeat() {
	ticker := time.NewTicker(b.config.HeartbeatTickInterval)
	defer ticker.Stop()

	for range ticker.C {
		b.inbox <- HeartbeatTickEvent{}
	}
}

func (b *Broker) handleSessionUpEvent(ev SessionUpEvent) {
	b.sessions[ev.CID] = ClientSession{
		Outbound:     ev.Outbound,
		AwaitingPong: false,
	}
}

func (b *Broker) handleSessionDownEvent(ev SessionDownEvent) {
	if session, ok := b.sessions[ev.CID]; ok {
		close(session.Outbound)
		delete(b.sessions, ev.CID)
	}
	b.registry.RemoveCID(ev.CID)
}

func (b *Broker) disconnectCID(cid int64, session ClientSession) {
	close(session.Outbound)
	delete(b.sessions, cid)
	b.registry.RemoveCID(cid)
}

func (b *Broker) handleCmdEvent(ev CmdEvent) {
	switch cmd := ev.Cmd.(type) {
	case codec.Ping:
		session, ok := b.sessions[ev.CID]
		if !ok {
			break
		}
		select {
		case session.Outbound <- codec.Pong{}:
		default:
			b.disconnectCID(ev.CID, session)
		}
	case codec.Pong:
		if session, ok := b.sessions[ev.CID]; ok {
			session.AwaitingPong = false
			b.sessions[ev.CID] = session
		}
	case codec.Connect:
		session, ok := b.sessions[ev.CID]
		if !ok {
			break
		}
		select {
		case session.Outbound <- codec.OK{}:
		default:
			b.disconnectCID(ev.CID, session)
		}
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
			session, ok := b.sessions[sub.CID]
			if !ok {
				continue
			}
			msg := codec.Msg{
				Subject: cmd.Subject,
				SID:     sub.SID,
				Payload: cmd.Payload,
			}
			select {
			case session.Outbound <- msg:
			default:
				b.disconnectCID(sub.CID, session)
			}
		}

	case codec.Unsub:
		b.registry.RemoveSub(ev.CID, cmd.SID)
	}
}

func (b *Broker) handleProtocolErrorEvent(ev ProtocolErrorEvent) {
	session, ok := b.sessions[ev.CID]
	if !ok {
		b.registry.RemoveCID(ev.CID)
		return
	}

	select {
	case session.Outbound <- codec.Err{Message: ev.Msg}:
	default:
	}

	b.disconnectCID(ev.CID, session)
}

func (b *Broker) handleHeartbeatTickEvent(ev HeartbeatTickEvent) {
	_ = ev
	now := time.Now()

	for cid, session := range b.sessions {
		if session.AwaitingPong {
			if now.Sub(session.PingSentAt) >= b.config.HeartbeatTimeout {
				b.disconnectCID(cid, session)
			}
			continue
		}

		select {
		case session.Outbound <- codec.Ping{}:
			session.AwaitingPong = true
			session.PingSentAt = now
			b.sessions[cid] = session
		default:
			b.disconnectCID(cid, session)
		}
	}
}
