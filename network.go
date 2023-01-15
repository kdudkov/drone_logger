package main

import (
	"context"
	"net"
	"time"
)

const idleTimeout = time.Second * 5

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
			conn := app.conn.Load()
			if conn == nil {
				continue
			}
			if _, err := conn.Write(data); err != nil {
				app.logger.AddLine(err.Error())
			}
		case <-app.ctx.Done():
			return
		}
	}
}

func (app *App) pinger() {
	ping := createMessage(0x0b, make([]byte, 8))
	app.commandCh <- ping
	periodical(app.ctx, time.Second*3, func() {
		app.commandCh <- ping
		app.commandCh <- createMessage(0xa, make([]byte, 15))
	})
}

func (app *App) reader() {
	buf := make([]byte, 4096)
	for app.ctx.Err() == nil {
		conn := app.conn.Load()
		if conn == nil {
			continue
		}
		n, err := conn.Read(buf)
		if err != nil {
			app.logger.AddLine(err.Error())
			continue
		}
		app.setActivity()
		msg := make([]byte, n)
		copy(msg, buf[:n])
		if err := app.processMessage(msg); err != nil {
			app.logger.AddLine(err.Error())
			continue
		}
		app.redraw()
	}
}

func (app *App) startConn() {
	conn, err := dial("192.168.0.1:50000")
	if err != nil {
		app.logger.AddLine(err.Error())
	}
	if old := app.conn.Swap(conn); old != nil {
		old.Close()
	}
	if app.closeTimer == nil {
		app.closeTimer = time.AfterFunc(idleTimeout, app.closeIdle)
	} else {
		app.closeTimer.Reset(idleTimeout)
	}
}

func (app *App) setActivity() {
	app.lastData = time.Now()

	if app.closeTimer == nil {
		app.closeTimer = time.AfterFunc(idleTimeout, app.closeIdle)
	} else {
		app.closeTimer.Reset(idleTimeout)
	}
}

func (app *App) closeIdle() {
	idle := time.Now().Sub(app.lastData)

	if idle >= idleTimeout {
		app.logger.AddLine("no data, reconnect")
		app.startConn()
	}
}

func dial(addr string) (*net.UDPConn, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	return net.DialUDP("udp", nil, a)
}
