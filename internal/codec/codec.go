package codec

import (
	"bufio"
	"errors"
	"io"
)

type Codec struct {
	brw *bufio.ReadWriter
}

const maxPayloadBytes int64 = 8 * 1024 * 1024

func NewCodec(rw io.ReadWriter) (*Codec, error) {
	if rw == nil {
		return nil, errors.New("nil read writer received")
	}

	brw, ok := rw.(*bufio.ReadWriter)
	if !ok {
		brw = bufio.NewReadWriter(
			bufio.NewReader(rw),
			bufio.NewWriter(rw),
		)
	}
	return &Codec{
		brw: brw,
	}, nil
}

type scratchSpace struct {
	Kind    Kind
	Subject []byte
	SID     []byte
	Msg     []byte
	nBytes  []byte
}

func (c *Codec) Decode() (Command, error) {
	ss := scratchSpace{}

	state := ST_START
	for {
		b, err := c.brw.ReadByte()
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
		case ST_CMD_PING:
			ss.Kind = KindPing
		case ST_CMD_PONG:
			ss.Kind = KindPong
		case ST_CMD_SUB:
			ss.Kind = KindSub
		case ST_CMD_PUB:
			ss.Kind = KindPub
		case ST_CMD_UNSUB:
			ss.Kind = KindUnsub
		case ST_SUB_SUBJECT, ST_SUB_SUBJECT_DOT, ST_SUB_SUBJECT_GT, ST_SUB_SUBJECT_STAR:
			ss.Subject = append(ss.Subject, b)
		case ST_SUB_SID:
			ss.SID = append(ss.SID, b)

		case ST_PUB_SUBJECT, ST_PUB_SUBJECT_DOT:
			ss.Subject = append(ss.Subject, b)
		case ST_PUB_NUM_BYTES:
			ss.nBytes = append(ss.nBytes, b)
		case ST_PUB_PAYLOAD:
			size, err := parseDigitsInt64(ss.nBytes)
			if err != nil {
				return nil, errors.New("bad payload")
			}
			if size > maxPayloadBytes {
				return nil, errors.New("payload too large")
			}

			ss.Msg = make([]byte, size)
			n, err := io.ReadFull(c.brw, ss.Msg)
			if err != nil {
				return nil, err
			}

			if int64(n) != size {
				return nil, errors.New("did not get full payload")
			}

			if c, err := c.brw.ReadByte(); c != '\r' || err != nil {
				return nil, errors.New("bad payload")
			}
			state = ST_CR_END

		case ST_UNSUB_SID:
			ss.SID = append(ss.SID, b)
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
		return Pub{
			Subject: ss.Subject,
			Len:     int64(len(ss.Msg)),
			Msg:     ss.Msg,
		}, nil
	case KindSub:
		sid, err := parseDigitsInt64(ss.SID)
		if err != nil {
			return nil, errors.New("bad sid")
		}
		return Sub{
			Subject: ss.Subject,
			SID:     sid,
		}, nil
	case KindUnsub:
		sid, err := parseDigitsInt64(ss.SID)
		if err != nil {
			return nil, errors.New("bad sid")
		}
		return Unsub{
			SID: sid,
		}, nil
	default:
		return nil, errors.New("kind not implemented")
	}
}

func parseDigitsInt64(bytes []byte) (int64, error) {
	if len(bytes) == 0 {
		return 0, errors.New("empty digits")
	}

	var n int64
	for _, b := range bytes {
		if b < '0' || b > '9' {
			return 0, errors.New("invalid digit")
		}
		d := int64(b - '0')
		if n > (9223372036854775807-d)/10 {
			return 0, errors.New("int64 overflow")
		}
		n = n*10 + d
	}
	return n, nil
}
