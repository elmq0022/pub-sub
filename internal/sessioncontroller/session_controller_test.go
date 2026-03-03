package sessioncontroller

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/elmq0022/pub-sub/internal/broker"
	"github.com/elmq0022/pub-sub/internal/codec"
)

func TestSessionControllerNextClientIDConcurrent(t *testing.T) {
	controller := NewSessionController(nil)

	const total = 128
	got := make(chan int64, total)

	var wg sync.WaitGroup
	wg.Add(total)

	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			got <- controller.nextClientID()
		}()
	}

	wg.Wait()
	close(got)

	seen := make(map[int64]bool, total)
	for cid := range got {
		if seen[cid] {
			t.Fatalf("duplicate cid allocated: %d", cid)
		}
		seen[cid] = true
	}

	for cid := int64(0); cid < total; cid++ {
		if !seen[cid] {
			t.Fatalf("missing cid allocation: %d", cid)
		}
	}
}

func TestReaderLoopClosesConnBeforeSendingSessionDown(t *testing.T) {
	conn := newTestConn(io.EOF)
	brokerInbox := make(chan broker.BrokerEvent)
	var downOnce sync.Once
	done := make(chan struct{})

	go func() {
		readerLoop(42, conn, brokerInbox, &downOnce)
		close(done)
	}()

	waitForClosed(t, conn.closed)

	select {
	case <-done:
		t.Fatal("readerLoop returned before broker consumed SessionDownEvent")
	default:
	}

	ev := waitForBrokerEvent(t, brokerInbox)
	down, ok := ev.(broker.SessionDownEvent)
	if !ok {
		t.Fatalf("expected SessionDownEvent, got %T", ev)
	}
	if down.CID != 42 {
		t.Fatalf("expected cid 42, got %d", down.CID)
	}

	waitForDone(t, done)
}

func TestWriterLoopClosesConnBeforeSendingSessionDown(t *testing.T) {
	conn := newTestConn(nil)
	brokerInbox := make(chan broker.BrokerEvent)
	outbound := make(chan codec.OutboundCommands)
	var downOnce sync.Once
	done := make(chan struct{})

	close(outbound)

	go func() {
		writerLoop(7, conn, brokerInbox, outbound, &downOnce)
		close(done)
	}()

	waitForClosed(t, conn.closed)

	select {
	case <-done:
		t.Fatal("writerLoop returned before broker consumed SessionDownEvent")
	default:
	}

	ev := waitForBrokerEvent(t, brokerInbox)
	down, ok := ev.(broker.SessionDownEvent)
	if !ok {
		t.Fatalf("expected SessionDownEvent, got %T", ev)
	}
	if down.CID != 7 {
		t.Fatalf("expected cid 7, got %d", down.CID)
	}

	waitForDone(t, done)
}

type testConn struct {
	closed   chan struct{}
	readErr  error
	closeMu  sync.Mutex
	isClosed bool
}

func newTestConn(readErr error) *testConn {
	return &testConn{
		closed:  make(chan struct{}),
		readErr: readErr,
	}
}

func (c *testConn) Read(_ []byte) (int, error) {
	if c.readErr != nil {
		return 0, c.readErr
	}

	<-c.closed
	return 0, net.ErrClosed
}

func (c *testConn) Write(p []byte) (int, error) {
	select {
	case <-c.closed:
		return 0, net.ErrClosed
	default:
		return len(p), nil
	}
}

func (c *testConn) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if !c.isClosed {
		close(c.closed)
		c.isClosed = true
	}
	return nil
}

func (c *testConn) LocalAddr() net.Addr  { return testAddr("local") }
func (c *testConn) RemoteAddr() net.Addr { return testAddr("remote") }
func (c *testConn) SetDeadline(_ time.Time) error {
	return nil
}
func (c *testConn) SetReadDeadline(_ time.Time) error {
	return nil
}
func (c *testConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

type testAddr string

func (a testAddr) Network() string { return "test" }
func (a testAddr) String() string  { return string(a) }

func waitForClosed(t *testing.T, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for connection close")
	}
}

func waitForBrokerEvent(t *testing.T, ch <-chan broker.BrokerEvent) broker.BrokerEvent {
	t.Helper()

	select {
	case ev := <-ch:
		return ev
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broker event")
		return nil
	}
}

func waitForDone(t *testing.T, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for goroutine completion")
	}
}
