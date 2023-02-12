package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"
)

func (app *App) ResetConn() {
	addr, _ := net.ResolveUDPAddr("udp", ":50000")
	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		panic(err)
	}

	if old := app.conn.Swap(conn); old != nil {
		old.Close()
	}
}

func (app *App) ListenUDP() error {
	app.ResetConn()
	buf := make([]byte, 65535)

	for {
		conn := app.conn.Load()
		if conn == nil {
			continue
		}
		n, addr, err := conn.ReadFrom(buf)
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
		oldAddr := app.addr.Swap(a)

		if oldAddr == nil || oldAddr.IP.String() != a.IP.String() {
			fmt.Printf("new client %s\n", addr)
		}

		code := msg[1]
		data := msg[3 : len(msg)-2]

		switch code {
		case 1:
			// ping
			fmt.Printf("01: %s\n", arr2str(data))
			if data[6]&4 > 0 {
				fmt.Println("cmd lock/unlock")
				app.WriteData(func(d *Data) {
					d.locked = !d.locked
				})
			}

			if data[6]&16 > 0 && app.data.air {
				fmt.Println("cmd land")
				app.WriteData(func(d *Data) {
					d.air = false
					d.alt = 0
				})
			}

			if data[6]&8 > 0 && !app.data.air {
				fmt.Println("cmd takeoff")
				app.WriteData(func(d *Data) {
					d.air = true
					d.locked = true
					d.alt = 2
				})
			}

			if c := data[3]; c != 0x80 {
				app.WriteData(func(d *Data) {
					d.yaw -= float64(int(c)-0x80) / 40
					fmt.Printf("yaw: %.2f\n", d.yaw)
				})
			}

			break
		case 0x7e:
			msg := createMessage(0xfe, str2byte("48462d584c2d584c303132533031312e3032312e3132323205"))
			app.ch <- msg
			break
		case 0x0a:
			msg := createMessage(0x8a, str2byte("584c30313253465878800c08"))
			app.ch <- msg
			break
		case 0x0b:
			break
		case 0x1b:
			break
		default:
			fmt.Printf("%.2x: %s\n", code, arr2str(data))
		}

	}
	return nil
}

func (app *App) Sender() {
	for msg := range app.ch {
		conn := app.conn.Load()
		addr := app.addr.Load()
		if conn == nil || addr == nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		conn.WriteTo(msg, addr)
	}
}

func (app *App) Crone() {
	var n uint64

	for range time.Tick(time.Millisecond * 200) {
		addr := app.addr.Load()
		if addr == nil {
			continue
		}
		if time.Now().Sub(app.lastTime) > time.Second*3 {
			fmt.Println("remove client")
			app.addr.Store(nil)
		}
		n++
		//fmt.Println(arr2str(msg))
		app.ch <- app.makeGyro()
		if n%4 == 0 {
			app.ch <- app.makeLL()
			app.ch <- app.makeBatt()
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

func str2byte(s string) []byte {
	r, _ := hex.DecodeString(s)
	return r
}
