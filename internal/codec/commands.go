package codec

type Kind uint8

const (
	KindPing Kind = iota
	KindPong
	KindConnect
	KindSub
	KindPub
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

type Pub struct{}

func (Pub) Kind() Kind { return KindPub }

type Sub struct{}

func (Sub) Kind() Kind { return KindSub }

// CONNECT {}
// PING
// PONG
// SUB <subject> <sid>\r\n
// PUB <subject> <#bytes>\r\n[payload]\r\n
