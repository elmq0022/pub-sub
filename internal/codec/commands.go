package codec

import (
	"bufio"
	"errors"
	"strconv"
)

type Kind uint8

const (
	KindPing Kind = iota
	KindPong
	KindConnect
	KindPub
	KindSub
	KindUnsub
	KindMsg
	KindOK
	KindErr
)

type Command interface {
	Kind() Kind
}

type OutboundCommands interface {
	Command
	EncodeTo(*bufio.Writer) error
}

type InboundCommands interface {
	Command
	IsInboundCommand()
}

type Ping struct{}

func (Ping) Kind() Kind { return KindPing }
func (Ping) IsInboundCommand() {}

func (Ping) EncodeTo(w *bufio.Writer) error {
	if w == nil {
		return errors.New("nil writer")
	}
	_, err := w.WriteString("PING\r\n")
	return err
}

type Pong struct{}

func (Pong) Kind() Kind { return KindPong }
func (Pong) IsInboundCommand() {}

func (Pong) EncodeTo(w *bufio.Writer) error {
	if w == nil {
		return errors.New("nil writer")
	}
	_, err := w.WriteString("PONG\r\n")
	return err
}

type Connect struct{}

func (Connect) Kind() Kind { return KindConnect }
func (Connect) IsInboundCommand() {}

type Sub struct {
	Subject []byte
	SID     int64
}

func (Sub) Kind() Kind { return KindSub }
func (Sub) IsInboundCommand() {}

type Pub struct {
	Subject []byte
	Len     int64
	Payload []byte
}

func (Pub) Kind() Kind { return KindPub }
func (Pub) IsInboundCommand() {}

type Unsub struct {
	SID int64
}

func (Unsub) Kind() Kind { return KindUnsub }
func (Unsub) IsInboundCommand() {}

// Msg is outbound-only and serialized by the writer actor as:
// MSG <subject> <sid> <#bytes>\r\n[payload]\r\n
type Msg struct {
	Subject []byte
	SID     int64
	Payload []byte
}

func (Msg) Kind() Kind { return KindMsg }

func (m Msg) EncodeTo(w *bufio.Writer) error {
	if w == nil {
		return errors.New("nil writer")
	}
	if len(m.Subject) == 0 {
		return errors.New("empty subject")
	}
	if m.SID < 0 {
		return errors.New("invalid sid")
	}

	if _, err := w.WriteString("MSG "); err != nil {
		return err
	}
	if _, err := w.Write(m.Subject); err != nil {
		return err
	}
	if err := w.WriteByte(' '); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.FormatInt(m.SID, 10)); err != nil {
		return err
	}
	if err := w.WriteByte(' '); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(len(m.Payload))); err != nil {
		return err
	}
	if _, err := w.WriteString("\r\n"); err != nil {
		return err
	}
	if _, err := w.Write(m.Payload); err != nil {
		return err
	}
	if _, err := w.WriteString("\r\n"); err != nil {
		return err
	}

	return nil
}

// OK is outbound-only and serialized as: +OK\r\n
type OK struct{}

func (OK) Kind() Kind { return KindOK }

func (OK) EncodeTo(w *bufio.Writer) error {
	if w == nil {
		return errors.New("nil writer")
	}
	_, err := w.WriteString("+OK\r\n")
	return err
}

// Err is outbound-only and serialized as: -ERR <message>\r\n
type Err struct {
	Message string
}

func (Err) Kind() Kind { return KindErr }

func (e Err) EncodeTo(w *bufio.Writer) error {
	if w == nil {
		return errors.New("nil writer")
	}
	if _, err := w.WriteString("-ERR "); err != nil {
		return err
	}
	if _, err := w.WriteString(e.Message); err != nil {
		return err
	}
	_, err := w.WriteString("\r\n")
	return err
}
