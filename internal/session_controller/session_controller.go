package sessioncontroller

import (
	"bufio"
	"io"
	"net"
	"sync"

	"github.com/elmq0022/pub-sub/internal/broker"
	"github.com/elmq0022/pub-sub/internal/codec"
)

type SessionController struct {
	brokerInbox chan<- broker.BrokerEvent
	nextCID     int64
}

func NewSessionController(brokerInbox chan<- broker.BrokerEvent) *SessionController {
	return &SessionController{
		brokerInbox: brokerInbox,
	}
}

func (s *SessionController) Start(conn net.Conn) {
	cid := s.nextCID
	s.nextCID++
	outbound := make(chan codec.OutboundCommands)
	var downOnce sync.Once

	// TODO: consider a handshake channel to coordinate start.
	go writerLoop(cid, conn, s.brokerInbox, outbound, &downOnce)
	s.brokerInbox <- broker.SessionUpEvent{
		CID:      cid,
		Outbound: outbound,
	}
	go readerLoop(cid, conn, s.brokerInbox, &downOnce)
}

func sendSessionDownOnce(cid int64, brokerInbox chan<- broker.BrokerEvent, once *sync.Once) {
	once.Do(func() {
		brokerInbox <- broker.SessionDownEvent{CID: cid}
	})
}

func readerLoop(cid int64, conn net.Conn, brokerInbox chan<- broker.BrokerEvent, downOnce *sync.Once) {
	c, err := codec.NewCodec(conn)
	if err != nil {
		sendSessionDownOnce(cid, brokerInbox, downOnce)
		return
	}

	defer func() {
		sendSessionDownOnce(cid, brokerInbox, downOnce)
		_ = conn.Close()
	}()

	for {
		cmd, err := c.Decode()
		if err != nil {
			return
		}

		brokerInbox <- broker.CmdEvent{
			CID: cid,
			Cmd: cmd,
		}
	}
}

func writerLoop(
	cid int64,
	conn net.Conn,
	brokerInbox chan<- broker.BrokerEvent,
	outbound <-chan codec.OutboundCommands,
	downOnce *sync.Once,
) {
	var w io.Writer = conn
	b, ok := w.(*bufio.Writer)
	if !ok {
		b = bufio.NewWriter(w)
	}

	defer func() {
		sendSessionDownOnce(cid, brokerInbox, downOnce)
		_ = conn.Close()
	}()

	for cmd := range outbound {
		err := cmd.EncodeTo(b)
		if err != nil {
			return
		}
		if err := b.Flush(); err != nil {
			return
		}
	}
}
