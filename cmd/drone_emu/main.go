package main

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type App struct {
	addr     atomic.Pointer[net.UDPAddr]
	conn     atomic.Pointer[net.UDPConn]
	lastTime time.Time
	ch       chan []byte
	data     *Data
	n        byte
	mx       sync.RWMutex
}

type Data struct {
	roll, pitch, yaw float64
	sats             byte
	lat, lon         float64
	alt, dist        float32
	hsp, vsp         float32
	locked           bool
	inAir            bool
	route            bool
	landing          bool
	hiSpeed          bool
}

func NewApp() *App {
	return &App{
		addr: atomic.Pointer[net.UDPAddr]{},
		ch:   make(chan []byte, 10),
		mx:   sync.RWMutex{},
		data: &Data{
			roll:   0,
			pitch:  0,
			yaw:    0,
			sats:   14,
			lat:    59.932970,
			lon:    30.234677,
			alt:    0,
			dist:   416,
			vsp:    0,
			locked: true,
			inAir:  false,
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
	go app.Crone()

	go app.ListenLwCmd()
	go app.ListenLwStream()

	app.ListenUDP()
}

func main() {
	NewApp().Run()
}
