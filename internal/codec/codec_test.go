package codec

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCodec(t *testing.T) {
	t.Run("nil read writer returns error", func(t *testing.T) {
		c, err := NewCodec(nil)
		require.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("wraps non buffered read writer", func(t *testing.T) {
		rw := &bytes.Buffer{}
		c, err := NewCodec(rw)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.brw)
	})

	t.Run("reuses existing bufio read writer", func(t *testing.T) {
		brw := bufio.NewReadWriter(
			bufio.NewReader(strings.NewReader("PING\r\n")),
			bufio.NewWriter(io.Discard),
		)
		c, err := NewCodec(brw)
		require.NoError(t, err)
		assert.Same(t, brw, c.brw)
	})
}

func TestCodecDecodeSuccess(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Command
	}{
		{name: "connect", input: "CONNECT {}\r\n", want: Connect{}},
		{name: "ping", input: "PING\r\n", want: Ping{}},
		{name: "pong", input: "PONG\r\n", want: Pong{}},
		{name: "sub", input: "SUB foo.bar 42\r\n", want: Sub{Subject: []byte("foo.bar"), SID: 42}},
		{name: "unsub", input: "UNSUB 9001\r\n", want: Unsub{SID: 9001}},
		{name: "pub", input: "PUB foo.bar 5\r\nhello\r\n", want: Pub{Subject: []byte("foo.bar"), Len: 5, Msg: []byte("hello")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCodec(bytes.NewBufferString(tt.input))
			require.NoError(t, err)

			got, err := c.Decode()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCodecDecodeErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		errText string
	}{
		{name: "bad parse", input: "BROKEN\r\n", errText: "bad parse"},
		{name: "bad payload digits", input: "PUB foo a\r\n", errText: "bad parse"},
		{name: "payload too large", input: "PUB foo 8388609\r\n", errText: "payload too large"},
		{name: "payload read short", input: "PUB foo 5\r\nhel", errText: "EOF"},
		{name: "payload missing trailing crlf", input: "PUB foo 3\r\nheyX", errText: "bad payload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCodec(bytes.NewBufferString(tt.input))
			require.NoError(t, err)

			_, err = c.Decode()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errText)
		})
	}
}

func TestCreateCmd(t *testing.T) {
	tests := []struct {
		name    string
		ss      scratchSpace
		want    Command
		errText string
	}{
		{name: "connect", ss: scratchSpace{Kind: KindConnect}, want: Connect{}},
		{name: "ping", ss: scratchSpace{Kind: KindPing}, want: Ping{}},
		{name: "pong", ss: scratchSpace{Kind: KindPong}, want: Pong{}},
		{name: "pub", ss: scratchSpace{Kind: KindPub, Subject: []byte("s"), Msg: []byte("abc")}, want: Pub{Subject: []byte("s"), Len: 3, Msg: []byte("abc")}},
		{name: "sub bad sid", ss: scratchSpace{Kind: KindSub, Subject: []byte("s"), SID: []byte("x")}, errText: "bad sid"},
		{name: "unsub bad sid", ss: scratchSpace{Kind: KindUnsub, SID: []byte("x")}, errText: "bad sid"},
		{name: "unknown kind", ss: scratchSpace{Kind: Kind(255)}, errText: "kind not implemented"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createCmd(tt.ss)
			if tt.errText != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errText)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDigitsInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    int64
		errText string
	}{
		{name: "single digit", input: []byte("7"), want: 7},
		{name: "multi digit", input: []byte("12345"), want: 12345},
		{name: "max int64", input: []byte("9223372036854775807"), want: 9223372036854775807},
		{name: "empty", input: nil, errText: "empty digits"},
		{name: "invalid", input: []byte("12a"), errText: "invalid digit"},
		{name: "overflow", input: []byte("9223372036854775808"), errText: "int64 overflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDigitsInt64(tt.input)
			if tt.errText != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errText)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeReturnsUnderlyingReadError(t *testing.T) {
	c, err := NewCodec(bytes.NewBufferString(""))
	require.NoError(t, err)

	_, err = c.Decode()
	require.Error(t, err)
	assert.True(t, errors.Is(err, io.EOF))
}

func TestCodecDecodeSequentialCommands(t *testing.T) {
	input := "PING\r\nPONG\r\nSUB foo 7\r\nPUB foo 5\r\nhello\r\nUNSUB 7\r\nCONNECT {}\r\n"
	c, err := NewCodec(bytes.NewBufferString(input))
	require.NoError(t, err)

	expected := []Command{
		Ping{},
		Pong{},
		Sub{Subject: []byte("foo"), SID: 7},
		Pub{Subject: []byte("foo"), Len: 5, Msg: []byte("hello")},
		Unsub{SID: 7},
		Connect{},
	}

	for i, want := range expected {
		got, decodeErr := c.Decode()
		require.NoError(t, decodeErr, "decode index %d", i)
		assert.Equal(t, want, got, "decode index %d", i)
	}

	_, err = c.Decode()
	require.Error(t, err)
	assert.True(t, errors.Is(err, io.EOF))
}

func TestCodecDecodePayloadBoundaries(t *testing.T) {
	t.Run("zero bytes payload", func(t *testing.T) {
		c, err := NewCodec(bytes.NewBufferString("PUB foo 0\r\n\r\n"))
		require.NoError(t, err)

		got, err := c.Decode()
		require.NoError(t, err)
		assert.Equal(t, Pub{Subject: []byte("foo"), Len: 0, Msg: []byte{}}, got)
	})

	t.Run("one byte payload", func(t *testing.T) {
		c, err := NewCodec(bytes.NewBufferString("PUB foo 1\r\na\r\n"))
		require.NoError(t, err)

		got, err := c.Decode()
		require.NoError(t, err)
		assert.Equal(t, Pub{Subject: []byte("foo"), Len: 1, Msg: []byte("a")}, got)
	})

	t.Run("max payload bytes", func(t *testing.T) {
		payload := bytes.Repeat([]byte("a"), int(maxPayloadBytes))
		var input bytes.Buffer
		_, _ = input.WriteString(fmt.Sprintf("PUB foo %d\r\n", maxPayloadBytes))
		_, _ = input.Write(payload)
		_, _ = input.WriteString("\r\n")

		c, err := NewCodec(&input)
		require.NoError(t, err)

		got, err := c.Decode()
		require.NoError(t, err)

		pub, ok := got.(Pub)
		require.True(t, ok)
		assert.Equal(t, []byte("foo"), pub.Subject)
		assert.Equal(t, maxPayloadBytes, pub.Len)
		require.Len(t, pub.Msg, int(maxPayloadBytes))
		assert.Equal(t, byte('a'), pub.Msg[0])
		assert.Equal(t, byte('a'), pub.Msg[len(pub.Msg)-1])
	})

	t.Run("max payload plus one rejected", func(t *testing.T) {
		c, err := NewCodec(bytes.NewBufferString(fmt.Sprintf("PUB foo %d\r\n", maxPayloadBytes+1)))
		require.NoError(t, err)

		_, err = c.Decode()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "payload too large")
	})
}

func TestCodecDecodeWithChunkedReader(t *testing.T) {
	rw := &chunkedReadWriter{
		data:  []byte("PUB foo 5\r\nhello\r\n"),
		chunk: 1,
	}
	c, err := NewCodec(rw)
	require.NoError(t, err)

	got, err := c.Decode()
	require.NoError(t, err)
	assert.Equal(t, Pub{Subject: []byte("foo"), Len: 5, Msg: []byte("hello")}, got)
}

func TestCodecDecodeLongRunMixedCommands(t *testing.T) {
	var input bytes.Buffer
	expected := make([]Command, 0, 1000)

	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			_, _ = input.WriteString("PING\r\n")
			expected = append(expected, Ping{})
		} else {
			msg := fmt.Sprintf("m%04d", i)
			subject := fmt.Sprintf("s%d", i)
			_, _ = input.WriteString(fmt.Sprintf("PUB %s %d\r\n%s\r\n", subject, len(msg), msg))
			expected = append(expected, Pub{
				Subject: []byte(subject),
				Len:     int64(len(msg)),
				Msg:     []byte(msg),
			})
		}
	}

	c, err := NewCodec(&input)
	require.NoError(t, err)

	for i, want := range expected {
		got, decodeErr := c.Decode()
		require.NoError(t, decodeErr, "decode index %d", i)
		assert.Equal(t, want, got, "decode index %d", i)
	}
}

func FuzzCodecDecodeDoesNotPanic(f *testing.F) {
	f.Add("PING\r\n")
	f.Add("PUB foo 3\r\nhey\r\n")
	f.Add("SUB foo.> 1\r\n")
	f.Add("UNSUB 9\r\n")
	f.Add("CONNECT {}\r\n")

	f.Fuzz(func(t *testing.T, input string) {
		c, err := NewCodec(bytes.NewBufferString(input))
		require.NoError(t, err)

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Decode panicked for input %q: %v", input, r)
			}
		}()

		_, _ = c.Decode()
	})
}

type chunkedReadWriter struct {
	data  []byte
	pos   int
	chunk int
	w     bytes.Buffer
}

func (c *chunkedReadWriter) Read(p []byte) (int, error) {
	if c.pos >= len(c.data) {
		return 0, io.EOF
	}
	n := c.chunk
	if n <= 0 {
		n = 1
	}
	remaining := len(c.data) - c.pos
	if n > remaining {
		n = remaining
	}
	if n > len(p) {
		n = len(p)
	}
	copy(p, c.data[c.pos:c.pos+n])
	c.pos += n
	return n, nil
}

func (c *chunkedReadWriter) Write(p []byte) (int, error) {
	return c.w.Write(p)
}
