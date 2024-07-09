package main

import (
	"drone/pkg/protocol"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type App struct {
	logger   *slog.Logger
	addr     atomic.Pointer[net.UDPAddr]
	conn     *net.UDPConn
	lastTime time.Time
	ch       chan *protocol.Message
	data     *Data
	mx       sync.RWMutex
}

type Data struct {
	batt             float64
	roll, pitch, yaw float64
	sats             byte
	lat, lon         float64
	alt              float64
	dist             int
	hsp, vsp         float64
	flags            map[string]bool
	cmd              []byte
}

func NewApp() *App {
	return &App{
		logger: slog.Default(),
		addr:   atomic.Pointer[net.UDPAddr]{},
		ch:     make(chan *protocol.Message, 10),
		mx:     sync.RWMutex{},
		data: &Data{
			batt:  8.9,
			roll:  0,
			pitch: 0,
			yaw:   0,
			sats:  14,
			lat:   59.932970,
			lon:   30.234677,
			alt:   0,
			dist:  416,
			vsp:   3.7,
			hsp:   -8.2,
			flags: map[string]bool{
				"in_air":   false,
				"armed":    false,
				"hi_speed": false,
				"sat_good": true,
			},
		},
	}
}

func (app *App) WriteData(f func(d *Data)) {
	app.mx.Lock()
	defer app.mx.Unlock()
	f(app.data)
}

func (app *App) Run() {
	go app.Sender()
	go app.Cron()

	go app.ListenLwCmd()
	go app.ListenLwStream()

	app.ListenUDP()
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	slog.SetDefault(logger)

	NewApp().Run()
}
