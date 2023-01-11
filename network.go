package main

import (
	"context"
	"net"
	"time"
)

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

func (app *App) sender4() {
	periodical(app.ctx, time.Second*3, func() {
		conn, err := net.Dial("udp", "192.168.0.1:40000")
		if err != nil {
			app.logger.AddLine(err.Error())
			return
		}
		err = conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
		if err != nil {
			app.logger.AddLine(err.Error())
			return
		}

		_, err = conn.Write([]byte{0x63, 0x63, 1, 0, 0, 0, 0})
		if err != nil {
			app.logger.AddLine(err.Error())
		}
		_ = conn.Close()
	})
}

func (app *App) sender5() {
	for {
		select {
		case data := <-app.commandCh:
			if _, err := app.conn.Write(data); err != nil {
				app.logger.AddLine(err.Error())
			}
		case <-app.ctx.Done():
			return
		}
	}
}

func (app *App) pinger() {
	ping := createMessage(0x0b, []byte{0, 0, 0, 0, 0, 0, 0, 0})
	app.commandCh <- ping
	periodical(app.ctx, time.Second*3, func() {
		app.commandCh <- ping
	})
}

func (app *App) reader() {
	buf := make([]byte, 4096)
	for app.ctx.Err() == nil {
		n, err := app.conn.Read(buf)
		if err != nil {
			app.logger.AddLine(err.Error())
			continue
		}
		app.lastData = time.Now()
		msg := make([]byte, n)
		copy(msg, buf[:n])
		if err := app.processMessage(msg); err != nil {
			app.logger.AddLine(err.Error())
			continue
		}
		app.redraw()
	}
}
