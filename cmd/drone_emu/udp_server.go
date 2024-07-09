package main

import (
	"drone/pkg/protocol"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"time"
)

func (app *App) Reconnect() {
	addr, _ := net.ResolveUDPAddr("udp", ":50000")
	conn, err := net.ListenUDP("udp4", addr)

	if err != nil {
		app.logger.Error("error", slog.Any("error", err))
		panic(err)
	}

	app.logger.Info("start listening :50000")
	app.conn = conn
}

func (app *App) ListenUDP() error {
	app.Reconnect()
	buf := make([]byte, 65535)

	for {
		n, addr, err := app.conn.ReadFromUDP(buf)

		if err != nil {
			app.logger.Error("read error", slog.Any("error", err))
			return err
		}

		m := &protocol.Message{}

		if err := m.Unmarshal(buf[:n]); err != nil {
			app.logger.Error("unmarshal error", slog.Any("error", err))
			continue
		}

		app.lastTime = time.Now()
		oldAddr := app.addr.Swap(addr)

		if oldAddr == nil || oldAddr.IP.String() != addr.IP.String() {
			app.logger.Info("new client " + addr.String())
		}

		switch m.Code {
		case 1:
			//app.logger.Debug("message " + m.String())
			app.processCommand(m.Data)
		case 0x7e:
			msg := protocol.NewMessage(0xfe, str2byte("48462d584c2d584c303132533031312e3032312e3132323205"))
			app.ch <- msg
		case 0x0a:
			msg := protocol.NewMessage(0x8a, str2byte("584c30313253465878800c08"))
			app.ch <- msg
		case 0x0b:
			break
		case 0x1b:
			break
		case 0x1e:
			break
		}
	}
}

func (app *App) Sender() {
	app.logger.Info("sender is started")

	for msg := range app.ch {
		addr := app.addr.Load()

		if app.conn == nil || addr == nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		msg.Back = true

		if _, err := app.conn.WriteTo(msg.Marshal(), addr); err != nil {
			app.logger.Error("write err", slog.Any("error", err))
		}
	}
}

func (app *App) Cron() {
	var n int

	for range time.Tick(time.Millisecond * 500) {
		addr := app.addr.Load()
		if addr == nil {
			continue
		}
		if time.Now().Sub(app.lastTime) > time.Second*3 {
			app.logger.Info("remove stale client " + addr.String())
			app.addr.Store(nil)
		}

		n++
		app.data.dist = n * 10 % 1500

		app.ch <- app.makeGyro(n)

		app.ch <- app.makeLL(n)
		app.ch <- app.makeBatt(n)
	}
}

func (app *App) processCommand(data []byte) {
	var oldData []byte

	app.WriteData(func(d *Data) {
		oldData = d.cmd
		d.cmd = data
	})

	for i, n := range oldData {
		if n != data[i] {
			fmt.Printf("%d %.2x -> %.2x\n", i, n, data[i])
		}
	}
	if data[6]&4 > 0 {
		app.WriteData(func(d *Data) {
			d.flags["armed"] = !d.flags["armed"]
			fmt.Printf("cmd arm/disarm, armed: %v\n", d.flags["armed"])
		})
	}

	if data[6]&8 > 0 && !app.data.flags["in_air"] {
		fmt.Println("cmd takeoff")
		app.WriteData(func(d *Data) {
			d.flags["in_air"] = true
			d.flags["armed"] = false
			d.alt = 8
		})
	}

	if data[6]&16 > 0 && app.data.flags["in_air"] {
		fmt.Println("cmd land")
		app.WriteData(func(d *Data) {
			d.flags["landing"] = true
		})

		time.AfterFunc(time.Second*3, func() {
			app.WriteData(func(d *Data) {
				d.flags["landing"] = false
				d.flags["in_air"] = false
				d.flags["armed"] = true
				d.alt = 0
			})
		})
	}

	if c := data[3]; c != 0x80 {
		app.WriteData(func(d *Data) {
			d.yaw -= float64(int(c)-0x80) / 40
			fmt.Printf("yaw: %.2f\n", d.yaw)
		})
	}

	if data[5] == 0x18 {
		app.WriteData(func(d *Data) {
			d.flags["hi_speed"] = true
		})
	}

	if data[5] == 8 {
		app.WriteData(func(d *Data) {
			d.flags["hi_speed"] = false
		})
	}

	if d := data[7]; d != 0 {
		fmt.Println("cmd gimble")
		if d == 0x40 {
			// gimble down
		}
		if d == 0x80 {
			// gimble up
		}
	}

	// 8 1 - ip, 0 - remote
}

func str2byte(s string) []byte {
	r, _ := hex.DecodeString(s)
	return r
}
