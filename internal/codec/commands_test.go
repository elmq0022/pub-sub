package codec

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutboundEncodeToWritesExpectedWireFormat(t *testing.T) {
	tests := []struct {
		name string
		cmd  OutboundCommands
		want string
	}{
		{
			name: "ping",
			cmd:  Ping{},
			want: "PING\r\n",
		},
		{
			name: "pong",
			cmd:  Pong{},
			want: "PONG\r\n",
		},
		{
			name: "msg with payload",
			cmd: Msg{
				Subject: []byte("foo.bar"),
				SID:     42,
				Payload: []byte("hello"),
			},
			want: "MSG foo.bar 42 5\r\nhello\r\n",
		},
		{
			name: "msg with empty payload",
			cmd: Msg{
				Subject: []byte("foo"),
				SID:     1,
				Payload: []byte{},
			},
			want: "MSG foo 1 0\r\n\r\n",
		},
		{
			name: "ok",
			cmd:  OK{},
			want: "+OK\r\n",
		},
		{
			name: "err",
			cmd:  Err{Message: "authorization violation"},
			want: "-ERR authorization violation\r\n",
		},
		{
			name: "err empty message",
			cmd:  Err{},
			want: "-ERR \r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			w := bufio.NewWriter(&out)

			err := tt.cmd.EncodeTo(w)
			require.NoError(t, err)
			require.NoError(t, w.Flush())

			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestOutboundEncodeToErrors(t *testing.T) {
	t.Run("nil writer", func(t *testing.T) {
		tests := []struct {
			name string
			cmd  OutboundCommands
		}{
			{name: "ping", cmd: Ping{}},
			{name: "pong", cmd: Pong{}},
			{name: "msg", cmd: Msg{Subject: []byte("foo"), SID: 1}},
			{name: "ok", cmd: OK{}},
			{name: "err", cmd: Err{Message: "boom"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.cmd.EncodeTo(nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "nil writer")
			})
		}
	})

	t.Run("msg empty subject", func(t *testing.T) {
		var out bytes.Buffer
		w := bufio.NewWriter(&out)

		err := (Msg{Subject: nil, SID: 7, Payload: []byte("hi")}).EncodeTo(w)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty subject")
	})

	t.Run("msg invalid sid", func(t *testing.T) {
		var out bytes.Buffer
		w := bufio.NewWriter(&out)

		err := (Msg{Subject: []byte("foo"), SID: -1, Payload: []byte("hi")}).EncodeTo(w)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sid")
	})
}
