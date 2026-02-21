package codec

type Kind uint8

const (
	KindPing Kind = iota
	KindPong
	KindConnect
	KindPub
	KindSub
	KindUnsub
)

type Command interface {
	Kind() Kind
}

type Ping struct{}

func (Ping) Kind() Kind { return KindPing }

type Pong struct{}

func (Pong) Kind() Kind { return KindPong }

type Connect struct{}

func (Connect) Kind() Kind { return KindConnect }

type Sub struct {
	Subject []byte
	SID     int64
}

func (Sub) Kind() Kind { return KindSub }

type Pub struct {
	Subject []byte
	Len     int64
	Msg     []byte
}

func (Pub) Kind() Kind { return KindPub }

type Unsub struct {
	SID int64
}

func (Unsub) Kind() Kind { return KindUnsub }

// CONNECT {}
// PING
// PONG
// SUB <subject> <sid>\r\n
// PUB <subject> <#bytes>\r\n[payload]\r\n
