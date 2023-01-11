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
	logger    *Logger
}

func NewApp() *App {
	app := &App{
		commandCh: make(chan []byte, 10),
		data:      sync.Map{},
		info:      &DroneInfo{info: sync.Map{}},
		logger:    NewLogger(),
	}
	app.logger.cb = app.redraw
	return app
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
			fmt.Fprintf(v, "roll: %s pitch: %s yaw: %s\n",
				formatGyro(app.info.getFloat("roll")),
				formatGyro(app.info.getFloat("pitch")),
				formatGyro(app.info.getFloat("yaw")))
			fmt.Fprintf(v, "em: %d battery: %.2fv\n", app.info.getByte("em"), app.info.getFloat("battery"))
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
			_, size := v.Size()
			for _, l := range app.logger.GetLines(size) {
				fmt.Fprintln(v, l)
			}
		}

		return nil
	})
}

func formatGyro(v float64) string {
	s := fmt.Sprintf("%7.2f", v)
	if v > -5 && v < 5 {
		return WithColors(s, FgGreen)
	}
	if v > -30 && v < 30 {
		return WithColors(s, FgYellow)
	}
	return WithColors(s, Bold, FgRed)
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
		panic(err)
	}

	app.ctx, app.cancel = context.WithCancel(context.Background())

	app.conn, err = net.Dial("udp", "192.168.0.1:50000")
	if err != nil {
		panic(err)
	}

	//go sender4(ctx)
	go app.sender5()
	go app.reader()

	go app.pinger()

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
