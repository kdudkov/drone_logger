package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
)

type App struct {
	g         *gocui.Gui
	commandCh chan []byte
	conn      net.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	lastData  time.Time
	data      sync.Map
	info      *DroneInfo
}

func NewApp() *App {
	return &App{
		commandCh: make(chan []byte, 10),
		data:      sync.Map{},
		info:      &DroneInfo{info: sync.Map{}},
	}
}

func (app *App) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("info", 0, 0, maxX/2-1, maxY/2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
	}
	if v, err := g.SetView("data", 0, maxY/2+1, maxX/2-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
	}
	if v, err := g.SetView("log", maxX/2, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
	}
	return nil
}

func (app *App) redraw() {
	app.g.Update(func(gui *gocui.Gui) error {
		if v, err := gui.View("info"); err == nil {
			v.Clear()
			fmt.Fprintf(v, "lat: %.6f lon: %.6f\n", app.info.getFloat("lat"), app.info.getFloat("lon"))
			fmt.Fprintf(v, "roll: %.2f pitch: %.2f yaw: %.2f\n", app.info.getFloat("roll"), app.info.getFloat("pitch"), app.info.getFloat("yaw"))
		}
		if v, err := gui.View("data"); err == nil {
			v.Clear()
			m := make(map[int]string)
			keys := make([]int, 0)
			app.data.Range(func(key, value any) bool {
				m[int(key.(byte))] = arr2str(value.([]byte))
				keys = append(keys, int(key.(byte)))
				return true
			})
			sort.Ints(keys)
			for _, k := range keys {
				fmt.Fprintf(v, "%.2x: %s\n", k, m[k])
			}
		}
		if v, err := gui.View("log"); err == nil {
			v.Clear()
		}

		return nil
	})
}

func (app *App) sender4() {
	ticker := time.NewTicker(time.Second * 3)
	for app.ctx.Err() == nil {
		select {
		case <-ticker.C:
			conn, err := net.Dial("udp", "192.168.0.1:40000")
			if err != nil {
				panic(err)
			}
			err = conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
			if err != nil {
				panic(err)
			}

			_, err = conn.Write([]byte{0x63, 0x63, 1, 0, 0, 0, 0})
			if err != nil {
				panic(err)
			}
			_ = conn.Close()
		case <-app.ctx.Done():
			break
		}
	}
	ticker.Stop()
}

func (app *App) sender5() {
	for {
		select {
		case data := <-app.commandCh:
			if _, err := app.conn.Write(data); err != nil {
				fmt.Println(err)
			}
		case <-app.ctx.Done():
			return
		}
	}
}

func (app *App) pinger() {
	ticker := time.NewTicker(time.Second * 3)
	for app.ctx.Err() == nil {
		select {
		case <-ticker.C:
			app.commandCh <- createMessage(0x0b, []byte{0, 0, 0, 0, 0, 0, 0, 0})
		case <-app.ctx.Done():
			break
		}
	}
	ticker.Stop()
}

func (app *App) reader() {
	buf := make([]byte, 4096)
	for app.ctx.Err() == nil {
		n, err := app.conn.Read(buf)
		if err != nil {
			continue
		}
		app.lastData = time.Now()
		msg := make([]byte, n)
		copy(msg, buf[:n])
		if err := app.processMessage(msg); err != nil {
			fmt.Println(err)
			continue
		}
		app.redraw()
	}
}

func (app *App) Run() {
	var err error

	app.g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer app.g.Close()

	app.g.SetManagerFunc(app.layout)

	if err := app.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, app.quit); err != nil {
		log.Panicln(err)
	}

	app.ctx, app.cancel = context.WithCancel(context.Background())

	app.conn, err = net.Dial("udp", "192.168.0.1:50000")
	if err != nil {
		panic(err)
	}

	//go sender4(ctx)
	go app.sender5()

	app.commandCh <- createMessage(0x0b, []byte{0, 0, 0, 0, 0, 0, 0, 0})
	go app.pinger()
	go app.reader()

	if err := app.g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	app.g.Close()
}

func (app *App) quit(g *gocui.Gui, v *gocui.View) error {
	if app.cancel != nil {
		app.cancel()
	}

	return gocui.ErrQuit
}

func main() {
	NewApp().Run()
}
