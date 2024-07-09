package main

import (
	"context"
	"drone/pkg/protocol"
	"fmt"

	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"

	"github.com/kdudkov/lwdrone/msg"
)

type App struct {
	g *gocui.Gui

	cancel context.CancelFunc

	data    sync.Map
	info    *DroneInfo
	logger  *Logger
	logfile *os.File
	drone   protocol.Drone

	photoMx sync.Mutex

	onceCreate *sync.Once

	lw *msg.Lwdrone
}

func NewApp() *App {
	app := &App{
		data:       sync.Map{},
		info:       NewInfo(),
		logger:     NewLogger(),
		onceCreate: new(sync.Once),
		lw:         msg.NewDrone(),
	}

	if err := app.lw.SetTime(); err != nil {
		panic(err)
	}

	app.logger.cb = app.redraw
	app.drone = protocol.NewSg907(app.setValue, app.newDataCb)
	app.drone.SetLogger(app.logger)
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

func (app *App) Run() {
	var err error

	app.g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer app.g.Close()

	app.g.SetManagerFunc(app.layout)

	if err := app.bindings(); err != nil {
		panic(err)
	}

	defer app.logfile.Close()

	var ctx context.Context

	ctx, app.cancel = context.WithCancel(context.Background())

	app.drone.Start(ctx)

	if err != nil {
		panic(err)
	}

	if err := app.g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func (app *App) quit(g *gocui.Gui, v *gocui.View) error {
	if app.cancel != nil {
		app.cancel()
	}

	app.drone.Stop()
	return gocui.ErrQuit
}

func (app *App) newDataCb(code byte, data []byte) {
	if app.info.getFloat("lat") != 0 && app.info.getFloat("lon") != 0 {
		app.onceCreate.Do(func() {
			var err error
			fname := time.Now().Format("20060102_150405.log")
			app.logfile, err = os.Create(fname)
			if err != nil {
				fmt.Println(err)
				app.logfile = nil
			}
		})
		if app.logfile != nil {
			_, _ = app.logfile.WriteString(fmt.Sprintf("%s;%.7f;%.7f;%.1f\n",
				time.Now().Format(time.RFC3339),
				app.info.getFloat("lat"),
				app.info.getFloat("lon"),
				app.info.getFloat("alt"),
			))
		}
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

func (app *App) photo(g *gocui.Gui, v *gocui.View) error {
	go func() {
		if !app.photoMx.TryLock() {
			return

		}

		defer app.photoMx.Unlock()

		p, err := app.lw.TakePicture()
		if err != nil {
			app.logger.AddLine(fmt.Sprintf("error: %s\n", err.Error()))
			return
		}
		fname := time.Now().Format("20060102_150405.jpg")
		f, err := os.Create(fname)
		if err != nil {
			app.logger.AddLine(err.Error())
			return
		}
		_, _ = f.Write(p.Data)
		f.Close()
		app.logger.AddLine("photo saved")

		ff, err := os.OpenFile("geo.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

		if err != nil {
			app.logger.AddLine(fmt.Sprintf("error: %s\n", err.Error()))
			return
		}

		ff.WriteString(fmt.Sprintf("%s\t%.7f\t%.7f\t%.1f\n",
			fname,
			app.info.getFloat("lat"),
			app.info.getFloat("lon"),
			app.info.getFloat("alt")))
		ff.Close()
	}()
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
