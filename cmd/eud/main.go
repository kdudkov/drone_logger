package main

import (
	"context"
	"drone/protocol"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
)

type App struct {
	g *gocui.Gui

	ctx    context.Context
	cancel context.CancelFunc

	data    sync.Map
	info    *DroneInfo
	logger  *Logger
	logfile *os.File
	drone   protocol.Drone
}

func NewApp() *App {
	app := &App{
		data:   sync.Map{},
		info:   &DroneInfo{info: sync.Map{}},
		logger: NewLogger(),
	}
	app.logger.cb = app.redraw
	app.drone = protocol.NewSg907(app.setValue, app.newDataCb)
	app.drone.SetWriter(app.logger)
	return app
}

func (app *App) setValue(key string, val any) {
	app.info.put(key, val)
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
			fmt.Fprintf(v, "lat: %.7f lon: %.7f\n", app.info.getFloat("lat"), app.info.getFloat("lon"))
			fmt.Fprintf(v, "roll: %s pitch: %s yaw: %f\n",
				formatGyro(app.info.getFloat("roll")),
				formatGyro(app.info.getFloat("pitch")),
				app.info.getFloat("yaw"))
			fmt.Fprintf(v, "alt: %.2f dist %.2f\n", app.info.getFloat("alt"), app.info.getFloat("dist"))
			fmt.Fprintf(v, "em: %d battery: %.2fv\n", app.info.getByte("em"), app.info.getFloat("battery"))
			fmt.Fprintf(v, "sat: %d, s1: %d s2:%d\n", app.info.getByte("sat"), app.info.getByte("s1"), app.info.getByte("s2"))
			fmt.Fprintf(v, "f1: %d f2: %d", app.info.getByte("f1"), app.info.getByte("f2"))
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
	if err := app.g.SetKeybinding("", 'w', gocui.ModNone, app.up); err != nil {
		panic(err)
	}
	if err := app.g.SetKeybinding("", 's', gocui.ModNone, app.down); err != nil {
		panic(err)
	}

	fname := time.Now().Format("20060102_150405.log")
	app.logfile, err = os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer app.logfile.Close()

	app.ctx, app.cancel = context.WithCancel(context.Background())

	app.drone.Start(app.ctx)

	if err != nil {
		panic(err)
	}

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

func (app *App) newDataCb(code byte, data []byte) {
	if app.info.getFloat("lat") != 0 && app.info.getFloat("lon") != 0 {
		app.logfile.WriteString(fmt.Sprintf("%s;%.7f;%.7f;%.1f\n",
			time.Now().Format(time.RFC3339),
			app.info.getFloat("lat"),
			app.info.getFloat("lon"),
			app.info.getFloat("alt"),
		))
	}
	app.data.Store(code, data)
	app.redraw()
}

func (app *App) up(g *gocui.Gui, v *gocui.View) error {
	app.drone.AddCommand(protocol.NewCommandBuilder().Up(127).Build())

	return nil
}

func (app *App) down(g *gocui.Gui, v *gocui.View) error {
	app.drone.AddCommand(protocol.NewCommandBuilder().Down(127).Build())

	return nil
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
