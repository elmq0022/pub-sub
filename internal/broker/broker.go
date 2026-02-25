package broker

import (
	"github.com/elmq0022/pub-sub/internal/codec"
	"github.com/elmq0022/pub-sub/internal/subjectregistry"
)

type CommandEnvelope struct {
	CID int64
	Cmd codec.Command
}

type ClientRef struct {
	CID      int64
	Outbound chan<- OutboundMsg
}

type OutboundKind uint8

const (
	OutboundKindMsg OutboundKind = iota
	OutboundKindPong
)

type OutboundMsg struct {
	Kind    OutboundKind
	Subject []byte
	SID     int64
	Payload []byte
}

type Broker struct {
	r       subjectregistry.Registry
	inbox   <-chan CommandEnvelope
	writers map[int64]chan<- OutboundMsg
}

func (b *Broker) Run() {
	for {
		msg := <-b.inbox
		cmd := msg.Cmd
		cid := msg.CID

		switch cmd.Kind() {
		case codec.KindConnect:
		case codec.KindPing:
			out, ok := b.writers[cid]
			if !ok {
				continue
			}
			out <- OutboundMsg{Kind: OutboundKindPong}
		case codec.KindPong:
		case codec.KindSub:
			subCmd, ok := cmd.(codec.Sub)
			if !ok {
				continue
			}

			b.r.AddSub(string(subCmd.Subject), subjectregistry.Sub{
				CID:    int64,
				SID:    int64,
				Client: *subjectregistry.client,
			})
		case codec.KindPub:
			pubCmd, ok := cmd.(codec.Pub)
			if !ok {
				continue
			}

			subs, err := b.r.Lookup(string(pubCmd.Subject))
			if err != nil {
				continue
			}
			for _, sub := range subs {
				out, ok := b.writers[sub.CID]
				if !ok {
					continue
				}

				out <- OutboundMsg{
					Kind:    OutboundKindMsg,
					Subject: pubCmd.Subject,
					SID:     sub.SID,
					Payload: pubCmd.Payload,
				}
			}

		case codec.KindUnsub:
			unsubCmd, ok := cmd.(codec.Unsub)
			if !ok {
			}
			b.r.RemoveSub(cid, unsubCmd.SID)
		default:
		}
	}
}
