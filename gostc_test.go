package gostc

import (
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

type testServer struct {
	t *testing.T

	addr string
	conn *net.UDPConn
	ch   chan []byte
}

func newTestServer(t *testing.T) *testServer {
	s := &testServer{t: t}
	u, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", u)
	if err != nil {
		t.Fatal(err)
	}
	s.conn = conn
	s.addr = conn.LocalAddr().String()
	s.ch = make(chan []byte)
	go func() {
		for {
			buf := make([]byte, 1000)
			n, _, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			s.ch <- buf[:n]
		}
	}()
	return s
}

func (s *testServer) Close() {
	s.conn.Close()
}

func (s *testServer) expect(msg string) {
	got := string(<-s.ch)
	if got != msg {
		s.t.Errorf("got message %q; want %q", got, msg)
	}
}

func newTestServerClient(t *testing.T) (*testServer, *Client) {
	s := newTestServer(t)
	c, err := NewClient(s.addr)
	if err != nil {
		t.Fatal(err)
	}
	return s, c
}

func cycle(seq []float64) func() float64 {
	i := 0
	return func() float64 {
		v := seq[i]
		i++
		if i >= len(seq) {
			i = 0
		}
		return v
	}
}

func TestCount(t *testing.T) {
	s, c := newTestServerClient(t)
	defer s.Close()

	c.Count("foo", 3, 1)
	s.expect("foo:3|c")

	c.Count("foo", 3, 0.5)
	s.expect("foo:3|c@0.5")

	c.Count("blah", -123.456, 1)
	s.expect("blah:-123.456|c")

	c.Inc("incme")
	s.expect("incme:1|c")

	c.randFloat = cycle([]float64{0.6, 0.4})
	c.CountProb("foo", 3, 0.5) // nothin
	c.CountProb("bar", 3, 0.5)
	s.expect("bar:3|c@0.5")

	c.IncProb("foo", 0.5)
	c.IncProb("bar", 0.5)
	s.expect("bar:1|c@0.5")
}

func TestTime(t *testing.T) {
	s, c := newTestServerClient(t)
	defer s.Close()

	c.Time("foo", 3*time.Second)
	s.expect("foo:3000|ms")
}

func TestGauge(t *testing.T) {
	s, c := newTestServerClient(t)
	defer s.Close()

	c.Gauge("foo", 123.456)
	s.expect("foo:123.456|g")
}

func TestSet(t *testing.T) {
	s, c := newTestServerClient(t)
	defer s.Close()

	c.Set("foo", []byte("hello"))
	s.expect("foo:hello|s")
}

func TestBufferedMaxSize(t *testing.T) {
	s := newTestServer(t)
	c, err := NewBufferedClient(s.addr, 100, 12, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for i := byte(0); i < 4; i++ {
		c.Set("a", []byte{'a' + i})
	}
	c.Close()
	s.expect("a:a|s\na:b|s")
	s.expect("a:c|s\na:d|s")
}

func TestBufferedMinFlush(t *testing.T) {
	s := newTestServer(t)
	c, err := NewBufferedClient(s.addr, 100, 100, 3*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	var once sync.Once
	ch := make(chan struct{})
	// Assigning flushHook here isn't racy because the the bufferAndSend
	// goroutine doesn't access flushHook unless there are messages in its
	// buffer, and we haven't sent any.
	c.flushHook = func() { once.Do(func() { close(ch) }) }
	for i := byte(0); i < 4; i++ {
		c.Set("a", []byte{'a' + i})
		if i == 1 {
			<-ch // wait for flush
		}
	}
	c.Close()
	s.expect("a:a|s\na:b|s")
	s.expect("a:c|s\na:d|s")
}

type nopWriteCloser struct{}

func (nopWriteCloser) Write(b []byte) (int, error) {
	return len(b), nil
}

func (nopWriteCloser) Close() error {
	return nil
}

var devNull = nopWriteCloser{}

func NewBenchClient() *Client {
	c, err := NewClient("localhost:12345")
	if err != nil {
		panic(err)
	}
	c.c = devNull
	return c
}

func BenchmarkCount(b *testing.B) {
	c := NewBenchClient()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Count("foo.bar", 123, 0.5)
	}
}

func BenchmarkInc(b *testing.B) {
	c := NewBenchClient()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc("foo.bar")
	}
}

func BenchmarkTime(b *testing.B) {
	c := NewBenchClient()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Time("foo.bar", time.Second)
	}
}

func BenchmarkGauge(b *testing.B) {
	c := NewBenchClient()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Gauge("foo.bar", 123.456)
	}
}

func BenchmarkSet(b *testing.B) {
	c := NewBenchClient()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("foo.bar", []byte("hello world"))
	}
}

func BenchmarkCountProb(b *testing.B) {
	rand.Seed(0)
	c := NewBenchClient()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CountProb("foo.bar", 123, 0.1)
	}
}

func BenchmarkIncProb(b *testing.B) {
	rand.Seed(0)
	c := NewBenchClient()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.IncProb("foo.bar", 0.1)
	}
}
