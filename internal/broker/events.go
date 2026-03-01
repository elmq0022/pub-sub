package broker

import "github.com/elmq0022/pub-sub/internal/codec"

type BrokerEvent interface{ isBrokerEvent() }

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

type HeartbeatTickEvent struct{}

func (HeartbeatTickEvent) isBrokerEvent() {}
