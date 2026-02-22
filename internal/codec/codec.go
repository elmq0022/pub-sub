package codec

import (
	"bufio"
	"errors"
	"io"
)

type Codec struct {
	rw io.ReadWriter
}

func NewCodec(rw io.ReadWriter) *Codec {
	return &Codec{
		rw: rw,
	}
}

type scratchSpace struct {
	Kind Kind
}

var ss = scratchSpace{}

func (c *Codec) Decode() (Command, error) {
	ss = scratchSpace{}

	if c.rw == nil {
		return nil, errors.New("got a nil read write")
	}

	brw, ok := c.rw.(*bufio.ReadWriter)
	if !ok {
		brw = bufio.NewReadWriter(
			bufio.NewReader(c.rw),
			bufio.NewWriter(c.rw),
		)
	}

	state := ST_START
	for {
		b, err := brw.ReadByte()
		if err != nil {
			return nil, err
		}

		state = transitionTable[state][b]

		switch state {
		case ST_ERROR:
			return nil, errors.New("bad parse")
		case ST_DONE:
			cmd, err := createCmd(ss)
			return cmd, err
		case ST_CMD_CONNECT:
			ss.Kind = KindConnect
			brw.Flush()
		case ST_CMD_PING:
			ss.Kind = KindPing
			brw.Flush()
		case ST_CMD_PONG:
			ss.Kind = KindPong
			brw.Flush()
		case ST_CMD_SUB:
			ss.Kind = KindSub
			brw.Flush()
		case ST_CMD_PUB:
			ss.Kind = KindPub
			brw.Flush()
		case ST_CMD_UNSUB:
			ss.Kind = KindUnsub
			brw.Flush()
		}
	}
}

func createCmd(ss scratchSpace) (Command, error) {
	switch ss.Kind {
	case KindConnect:
		return Connect{}, nil
	case KindPing:
		return Ping{}, nil
	case KindPong:
		return Pong{}, nil
	case KindPub:
		return Pub{}, nil
	case KindSub:
		return Sub{}, nil
	case KindUnsub:
		return Unsub{}, nil
	default:
		return nil, errors.New("kind not implemented")
	}
}
