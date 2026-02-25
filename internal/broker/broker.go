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
