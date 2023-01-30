package main

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

type App struct {
	addr     atomic.Pointer[net.UDPAddr]
	lastTime time.Time
	ch       chan []byte
}

func NewApp() *App {
	return &App{
		addr: atomic.Pointer[net.UDPAddr]{},
		ch:   make(chan []byte, 10),
	}
}

func (app *App) Run() {
	go app.Sender()
	go app.Crone()

	app.ListenUDP()
}

func (app *App) ListenUDP() error {
	p, err := net.ListenPacket("udp", ":50000")

	if err != nil {
		panic(err)
	}

	buf := make([]byte, 65535)

	for {
		n, addr, err := p.ReadFrom(buf)
		if err != nil {
			fmt.Printf("read error: %v\n", err)
			return err
		}

		msg := buf[:n]

		if len(msg) < 3 || len(msg)-4 != int(msg[2]) {
			fmt.Printf("invalid lenght")
			continue
		}

		var csum byte
		for _, c := range msg[1 : len(msg)-1] {
			csum ^= c
		}

		if csum != msg[len(msg)-1] {
			fmt.Printf("invalid checksum: %.2x %.2x\n", csum, msg[len(msg)-1])
			continue
		}

		a := addr.(*net.UDPAddr)
		app.lastTime = time.Now()
		fmt.Println()
		oldAddr := app.addr.Swap(a)

		if oldAddr == nil || oldAddr.IP.String() != a.IP.String() {
			fmt.Printf("new client %s\n", addr)
		}

		//data := msg[3 : len(msg)-2]
		code := msg[1]
		fmt.Println(code)

	}
	return nil
}

func (app *App) Sender() {
	for msg := range app.ch {
		addr := app.addr.Load()
		if addr == nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			fmt.Printf("dial error: %v\n", err)
			continue
		}

		conn.Write(msg)
	}
}

func (app *App) Crone() {
	ticker := time.NewTicker(time.Millisecond * 200)

	for {
		select {
		case <-ticker.C:
			addr := app.addr.Load()
			if addr == nil {
				continue
			}
			if time.Now().Sub(app.lastTime) > time.Second*3 {
				fmt.Println("remove client")
				app.addr.Store(nil)
			}
			msg := createMessage(0x8b, []byte{0x2d, 0x01, 0x78, 0xff, 0x4a, 0x1c, 0xfb, 0x80, 0x26, 0x00, 0x00, 0x00, 0x00, 0x07})
			fmt.Println(arr2str(msg))
			app.ch <- msg
		}
	}
}

func createMessage(code byte, data []byte) []byte {
	res := make([]byte, len(data)+4)
	res[0] = 0x58
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

func arr2str(d []byte) string {
	b := strings.Builder{}
	for i, c := range d {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(fmt.Sprintf("%.2x", c))
	}
	return b.String()
}

func main() {
	NewApp().Run()
}
