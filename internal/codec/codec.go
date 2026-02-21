package codec

import (
	"bufio"
	"bytes"
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

func (c *Codec) Decode() (Command, error) {
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

	line, err := brw.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	fields := bytes.Fields(line)

	switch string(fields[0]) {
	case "PING":
		return Ping{}, nil
	case "PONG":
		return Pong{}, nil
	case "CONNECT":
		return Connect{}, nil
	case "PUB":
		return Pub{}, nil
	case "SUB":
		return Sub{}, nil
	default:
		return nil, errors.New("no command")
	}
}
