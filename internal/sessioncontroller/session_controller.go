package sessioncontroller

import (
	"bufio"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elmq0022/pub-sub/internal/broker"
	"github.com/elmq0022/pub-sub/internal/codec"
)

type SessionController struct {
	brokerInbox chan<- broker.BrokerEvent
	nextCID     atomic.Int64
}

func NewSessionController(brokerInbox chan<- broker.BrokerEvent) *SessionController {
	return &SessionController{
		brokerInbox: brokerInbox,
	}
}

func (s *SessionController) Start(conn net.Conn) {
	cid := s.nextCID.Add(1) - 1
	outbound := make(chan codec.OutboundCommands, 256)
	var downOnce sync.Once

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
		_ = conn.Close()
		sendSessionDownOnce(cid, brokerInbox, downOnce)
	}()

	for {
		cmd, err := c.Decode()
		if err != nil {
			if shouldEmitProtocolError(err) {
				brokerInbox <- broker.ProtocolErrorEvent{
					CID: cid,
					Msg: "unparsable command",
				}
			}
			return
		}

		brokerInbox <- broker.CmdEvent{
			CID: cid,
			Cmd: cmd,
		}
	}
}

func shouldEmitProtocolError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed) {
		return false
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return false
	}

	return true
}

func writerLoop(
	cid int64,
	conn net.Conn,
	brokerInbox chan<- broker.BrokerEvent,
	outbound <-chan codec.OutboundCommands,
	downOnce *sync.Once,
) {
	b := bufio.NewWriterSize(conn, 32*1024)

	defer func() {
		_ = conn.Close()
		sendSessionDownOnce(cid, brokerInbox, downOnce)
	}()

	const timeout = 5 * time.Second
	for cmd := range outbound {
		_ = conn.SetWriteDeadline(time.Now().Add(timeout))
		if err := cmd.EncodeTo(b); err != nil {
			return
		}
		if err := b.Flush(); err != nil {
			return
		}
	}
}
