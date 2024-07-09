package protocol

import (
	"context"
	"io"
	"net"
	"os"
	"sync/atomic"
	"time"
)

const idleTimeout = time.Second * 5

type Sg907 struct {
	cancel     context.CancelFunc
	commandCh  chan *Message
	conn       atomic.Pointer[net.UDPConn]
	lastData   time.Time
	closeTimer *time.Timer
	dataAdder  func(k string, val any)
	dataCb     func(code byte, data []byte)
	logger     io.Writer
	logAll     bool
	logFile    io.WriteCloser
}

func NewSg907(dataAdder func(k string, val any), dataCb func(code byte, data []byte)) *Sg907 {
	return &Sg907{
		dataAdder: dataAdder,
		dataCb:    dataCb,
		commandCh: make(chan *Message, 50),
		logAll:    true,
	}
}

func (s *Sg907) SetLogger(logger io.Writer) {
	s.logger = logger
}

func (s *Sg907) Start(ctx context.Context) {
	if s.logAll {
		f, err := os.Create(time.Now().Format("data_20060102_150405.log"))

		if err == nil {
			s.logFile = f
		} else {
			s.logStr(err.Error())
		}
	}

	ctx, s.cancel = context.WithCancel(ctx)

	s.startConn5k()
	//go sender40k(ctx)
	go s.sender50k(ctx)
	go s.reader50k(ctx)

	go s.pinger(ctx)
}

func (s *Sg907) Stop() {
	if s.logFile != nil {
		s.logFile.Close()
	}

	s.closeConn()
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Sg907) AddCommand(m *Message) {
	if m != nil && !m.Empty() {
		s.commandCh <- m
	}
}

func (s *Sg907) logStr(l string) {
	if s.logger != nil {
		_, _ = s.logger.Write([]byte(l))
	}
}

func periodical(ctx context.Context, t time.Duration, f func()) {
	ticker := time.NewTicker(t)
	defer ticker.Stop()

	for ctx.Err() == nil {
		select {
		case <-ticker.C:
			f()
		case <-ctx.Done():
			break
		}
	}
}

func (s *Sg907) sender40k(ctx context.Context) {
	periodical(ctx, time.Second*3, func() {
		conn, err := net.Dial("udp", "192.168.0.1:40000")
		if err != nil {
			s.logStr(err.Error())
			return
		}
		err = conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
		if err != nil {
			s.logStr(err.Error())
			return
		}

		_, err = conn.Write([]byte{0x63, 0x63, 1, 0, 0, 0, 0})
		if err != nil {
			s.logStr(err.Error())
		}
		_ = conn.Close()
	})
}

func (s *Sg907) sender50k(ctx context.Context) {
	for ctx.Err() == nil {
		select {
		case cmd := <-s.commandCh:
			conn := s.conn.Load()
			if conn == nil {
				continue
			}
			if _, err := conn.Write(cmd.Marshal()); err != nil {
				s.logStr(err.Error())
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Sg907) pinger(ctx context.Context) {
	s.commandCh <- NewMessage(0x7e, str2byte("48462d58582d5858585858583030302e3030302e30303030000000000000"))
	s.commandCh <- NewMessage(0xa, make([]byte, 15))
	ping := NewMessage(0x0b, make([]byte, 8))
	s.commandCh <- ping

	periodical(ctx, time.Second, func() {
		s.commandCh <- ping
	})
}

func (s *Sg907) reader50k(ctx context.Context) {
	buf := make([]byte, 4096)

	for ctx.Err() == nil {
		conn := s.conn.Load()

		if conn == nil {
			continue
		}

		n, err := conn.Read(buf)

		if err != nil {
			s.logStr(err.Error())
			continue
		}

		s.setActivity(true)

		if s.logAll {

		}

		m := &Message{}

		if err := m.Unmarshal(buf[:n]); err != nil {
			s.logStr(err.Error())
		}

		s.processMessage(m)
	}
}

func (s *Sg907) startConn5k() {
	conn, err := dial("192.168.0.1:50000")
	if err != nil {
		s.logStr(err.Error())
	}

	if old := s.conn.Swap(conn); old != nil {
		old.Close()
	}

	s.setActivity(false)
}

func (s *Sg907) setActivity(withMsg bool) {
	if withMsg {
		s.lastData = time.Now()
	}

	if s.closeTimer == nil {
		s.closeTimer = time.AfterFunc(idleTimeout, s.closeIdle)
	} else {
		s.closeTimer.Reset(idleTimeout)
	}
}

func (s *Sg907) closeConn() {
	if conn := s.conn.Swap(nil); conn != nil {
		_ = conn.Close()
	}
}

func (s *Sg907) closeIdle() {
	idle := time.Now().Sub(s.lastData)

	if idle >= idleTimeout {
		s.logStr("no data, reconnect")
		s.closeConn()
		s.startConn5k()
	}
}

func dial(addr string) (*net.UDPConn, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	return net.DialUDP("udp", nil, a)
}

func (s *Sg907) processMessage(msg *Message) {
	switch msg.Code {

	case 0x8b: // 14: 3900 f300 7e3d fb00290000 00 0007
		parse8b(msg.Data, s.dataAdder)

	case 0x8c: // 13: 00000000 00000000 e803 0000 00
		parse8c(msg.Data, s.dataAdder)

	case 0x8f: // 8: c0 f00c 0000010000
		parse8f(msg.Data, s.dataAdder)

	case 0xfe:
		s.dataAdder("version", string(msg.Data))
	}
	// c8 - 18 cc - 19 c4 - 17

	s.dataCb(msg.Code, msg.Data)
}
