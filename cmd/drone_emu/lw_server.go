package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kdudkov/lwdrone/msg"
)

type Streamer struct {
	conn      *net.TCPConn
	ch        chan *msg.Command
	streaming uint32
}

func NewStreamer(conn *net.TCPConn) *Streamer {
	return &Streamer{
		conn: conn,
		ch:   make(chan *msg.Command, 50),
	}
}

func (s *Streamer) start() {
	go s.sender()
	go s.reader()
}

func (s *Streamer) sender() {
	for cmd := range s.ch {
		if cmd == nil {
			return
		}
		if _, err := s.conn.Write(cmd.ToByte()); err != nil {
			s.conn.Close()
			return
		}
	}
}

func (s *Streamer) reader() {
	wg := new(sync.WaitGroup)
	defer func() {
		atomic.StoreUint32(&s.streaming, 0)
		wg.Wait()
		close(s.ch)
	}()

	buf := make([]byte, 65535)
	for {
		n, err := s.conn.Read(buf)
		if err != nil {
			return
		}

		cmd, err := msg.FromByte(buf[:n])
		if err != nil {
			return
		}

		if cmd.GetCode() == msg.CmdHeartbeat {
			continue
		}

		if cmd.GetCode() == msg.CmdStartstream {
			if atomic.CompareAndSwapUint32(&s.streaming, 0, 1) {
				wg.Add(1)
				go s.stream(wg)
			}
			continue
		}

		if cmd.GetCode() == msg.CmdStopstream {
			if atomic.CompareAndSwapUint32(&s.streaming, 1, 0) {
			}
			continue
		}

		ans := msg.NewCommand(cmd.GetCode(), nil)
		s.ch <- ans
	}
}

func (s *Streamer) stream(wg *sync.WaitGroup) {
	defer wg.Done()
	var c uint64
	for atomic.LoadUint32(&s.streaming) == 1 {
		c++
		n := 8000
		data := make([]byte, n)
		le := binary.LittleEndian

		le.PutUint32(data[4:], uint32(n-32))
		le.PutUint64(data[8:], c)

		msg := msg.NewCommand(msg.CmdRetstream, data)
		msg.SetStreamType(129)
		s.ch <- msg
		time.Sleep(time.Millisecond * 100)
	}
}

func (app *App) ListenLwCmd() {
	addr, _ := net.ResolveTCPAddr("tcp", ":8060")
	l, err := net.ListenTCP("tcp", addr)

	if err != nil {
		panic(err)
	}

	for {
		if conn, err := l.AcceptTCP(); err == nil {
			go app.commandSession(conn)
		}

	}
}

func (app *App) ListenLwStream() {
	addr, _ := net.ResolveTCPAddr("tcp", ":7060")
	l, err := net.ListenTCP("tcp", addr)

	if err != nil {
		panic(err)
	}

	for {
		if conn, err := l.AcceptTCP(); err == nil {
			NewStreamer(conn).start()
		}
	}
}

func (app *App) commandSession(conn *net.TCPConn) {
	defer conn.Close()

	buf := make([]byte, 65535)
	for {
		conn.SetDeadline(time.Now().Add(time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		cmd, err := msg.FromByte(buf[:n])
		if err != nil {
			return
		}

		fmt.Println(cmd)
		ans := msg.NewCommand(cmd.GetCode(), nil)
		conn.Write(ans.ToByte())
	}
}
