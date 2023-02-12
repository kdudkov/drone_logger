package protocol

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"
)

const idleTimeout = time.Second * 5

type Sg907 struct {
	ctx        context.Context
	cancel     context.CancelFunc
	commandCh  chan []byte
	conn       atomic.Pointer[net.UDPConn]
	connected  int32
	lastData   time.Time
	closeTimer *time.Timer
	dataAdder  func(k string, val any)
	dataCb     func(code byte, data []byte)
	writer     io.Writer
}

func NewSg907(dataAdder func(k string, val any), dataCb func(code byte, data []byte)) *Sg907 {
	return &Sg907{
		dataAdder: dataAdder,
		dataCb:    dataCb,
		commandCh: make(chan []byte, 50),
	}
}

func (s *Sg907) SetWriter(writer io.Writer) {
	s.writer = writer
}

func (s *Sg907) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.startConn()
	//go sender4(ctx)
	go s.sender5()
	go s.reader()

	go s.pinger()
}

func (s *Sg907) AddCommand(b []byte) {
	s.commandCh <- b
}

func (s *Sg907) addLine(l string) {
	s.writer.Write([]byte(l))
}

func periodical(ctx context.Context, t time.Duration, f func()) {
	ticker := time.NewTicker(t)
	for ctx.Err() == nil {
		select {
		case <-ticker.C:
			f()
		case <-ctx.Done():
			break
		}
	}
	ticker.Stop()
}

func (s *Sg907) sender4() {
	periodical(s.ctx, time.Second*3, func() {
		conn, err := net.Dial("udp", "192.168.0.1:40000")
		if err != nil {
			s.addLine(err.Error())
			return
		}
		err = conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
		if err != nil {
			s.addLine(err.Error())
			return
		}

		_, err = conn.Write([]byte{0x63, 0x63, 1, 0, 0, 0, 0})
		if err != nil {
			s.addLine(err.Error())
		}
		_ = conn.Close()
	})
}

func (s *Sg907) sender5() {
	for {
		select {
		case data := <-s.commandCh:
			conn := s.conn.Load()
			if conn == nil {
				continue
			}
			if _, err := conn.Write(data); err != nil {
				s.addLine(err.Error())
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Sg907) pinger() {
	s.commandCh <- createMessage(0x7e, str2byte("48462d58582d5858585858583030302e3030302e30303030000000000000"))
	s.commandCh <- createMessage(0xa, make([]byte, 15))
	ping := createMessage(0x0b, make([]byte, 8))
	s.commandCh <- ping

	periodical(s.ctx, time.Second, func() {
		s.commandCh <- ping
	})
}

func (s *Sg907) reader() {
	buf := make([]byte, 4096)
	for s.ctx.Err() == nil {
		conn := s.conn.Load()
		if conn == nil {
			continue
		}
		n, err := conn.Read(buf)
		if err != nil {
			s.addLine(err.Error())
			continue
		}
		s.setActivity(true)
		msg := make([]byte, n)
		copy(msg, buf[:n])
		if err := s.processMessage(msg); err != nil {
			s.addLine(err.Error())
			continue
		}
	}
}

func (s *Sg907) startConn() {
	conn, err := dial("192.168.0.1:50000")
	if err != nil {
		s.addLine(err.Error())
	}
	if old := s.conn.Swap(conn); old != nil {
		old.Close()
	}
	s.setActivity(false)
}

func (s *Sg907) setActivity(withMsg bool) {
	if withMsg {
		s.lastData = time.Now()
		atomic.StoreInt32(&s.connected, 1)
	}
	if s.closeTimer == nil {
		s.closeTimer = time.AfterFunc(idleTimeout, s.closeIdle)
	} else {
		s.closeTimer.Reset(idleTimeout)
	}
}

func (s *Sg907) closeIdle() {
	idle := time.Now().Sub(s.lastData)

	if idle >= idleTimeout {
		s.addLine("no data, reconnect")
		atomic.StoreInt32(&s.connected, 0)
		s.startConn()
	}
}

func dial(addr string) (*net.UDPConn, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	return net.DialUDP("udp", nil, a)
}

func (s *Sg907) processMessage(msg []byte) error {
	if len(msg) < 3 || len(msg)-4 != int(msg[2]) {
		return fmt.Errorf("invalid lenght")
	}

	var csum byte
	for _, c := range msg[1 : len(msg)-1] {
		csum ^= c
	}

	if csum != msg[len(msg)-1] {
		return fmt.Errorf("invalid checksum: %.2x %.2x", csum, msg[len(msg)-1])
	}

	data := msg[3 : len(msg)-2]

	le := binary.LittleEndian

	switch msg[1] {
	case 0x8b: // 14: 3900 f300 7e3d fb00290000 00 0007
		s.dataAdder("roll", float64(int16(le.Uint16(data[0:2])))/100)
		s.dataAdder("pitch", float64(int16(le.Uint16(data[2:4])))/100)
		s.dataAdder("yaw", float64(int16(le.Uint16(data[4:6])))/100)
		s.dataAdder("locked", data[9]&(1<<3) == 0)
		s.dataAdder("in_air", data[9]&(1<<4) > 0)
		if data[9]&(1<<5) > 0 {
			s.dataAdder("state", "returning")
		}
		if data[9]&(1<<7) > 0 {
			s.dataAdder("state", "route")
		}
		s.dataAdder("sat", (data[11]&0x7c)>>2)
		s.dataAdder("sat_good", data[11]&128 > 0)
		if data[11]&1 > 0 {
			s.dataAdder("state", "landing")
		}
		if len(data) > 13 {
			s.dataAdder("hi_speed", data[13]&2 > 0)
			s.dataAdder("remote", data[13]&4 == 0)
		}

	case 0x8c: // 13: 00000000 00000000 e803 0000 00
		s.dataAdder("lon", float64(int32(le.Uint32(data[0:4])))/10000000)
		s.dataAdder("lat", float64(int32(le.Uint32(data[4:8])))/10000000)
		s.dataAdder("alt", float64(int16(le.Uint16(data[8:10])))/10-100)
		s.dataAdder("dist", float64(int16(le.Uint16(data[10:12])))*1.6)
		if len(data) > 12 {
			s.dataAdder("vsp", -float64(int16(data[12]))/10)
		}
	case 0x8f: // 8: c0 f00c 0000010000
		s.dataAdder("em", int(data[0])+int((data[1]&0xc0)>>6)*256)
		s.dataAdder("battery", float64((le.Uint16(data[1:3])&0x3ffc)>>2)/100)
		s.dataAdder("optical", data[4]&1 > 0)
		s.dataAdder("dist2", le.Uint16(data[5:7])>>1)
	case 0xfe:
		s.dataAdder("version", string(data))
	}
	// c8 - 18 cc - 19 c4 - 17

	s.dataCb(msg[1], data)
	return nil
}

func createMessage(code byte, data []byte) []byte {
	res := make([]byte, len(data)+4)
	res[0] = 0x68
	res[1] = code
	res[2] = byte(len(data))
	for i, c := range data {
		res[i+3] = c
	}
	var csum byte
	for _, c := range res[1 : len(res)-1] {
		csum ^= c
	}
	res[len(res)-1] = csum
	return res
}
