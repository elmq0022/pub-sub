package codec

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
	Payload []byte
}

func (Pub) Kind() Kind { return KindPub }

type Unsub struct {
	SID int64
}

func (Unsub) Kind() Kind { return KindUnsub }

// Msg is outbound-only and serialized by the writer actor as:
// MSG <subject> <sid> <#bytes>\r\n[payload]\r\n
type Msg struct {
	Subject []byte
	SID     int64
	Payload []byte
}

func (Msg) Kind() Kind { return KindMsg }

// OK is outbound-only and serialized as: +OK\r\n
type OK struct{}

func (OK) Kind() Kind { return KindOK }

// Err is outbound-only and serialized as: -ERR <message>\r\n
type Err struct {
	Message string
}

func (Err) Kind() Kind { return KindErr }
