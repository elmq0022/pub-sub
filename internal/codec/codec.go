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
	Kind    Kind
	Subject []byte
	SID     []byte
	Msg     []byte
	nBytes  []byte
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
			size := bytesToInt64(ss.nBytes)
			if size < 2 {
				return nil, errors.New("bad payload")
			}

			ss.Msg = make([]byte, size)
			n, err := io.ReadFull(brw, ss.Msg)
			if err != nil {
				return nil, err
			}

			if int64(n) != size {
				return nil, errors.New("did not get full payload")
			}

			ss.Msg = ss.Msg[:len(ss.Msg)]

			if c, err := brw.ReadByte(); c != '\r' || err != nil {
				return nil, errors.New("bad payload")
			}
			if c, err := brw.ReadByte(); c != '\n' || err != nil {
				return nil, errors.New("bad payload")
			}

			state = ST_DONE

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
			Len:     int64(len(ss.Subject)),
			Msg:     ss.Msg,
		}, nil
	case KindSub:
		return Sub{
			Subject: ss.Subject,
			SID:     bytesToInt64(ss.SID),
		}, nil
	case KindUnsub:
		return Unsub{
			SID: bytesToInt64(ss.SID),
		}, nil
	default:
		return nil, errors.New("kind not implemented")
	}
}

// NOTE: transition table ensures the bytes are all digits
func bytesToInt64(bytes []byte) int64 {
	value := int64(0)
	for _, b := range bytes {
		value *= 10
		value += int64(b - '0')
	}
	return value
}
