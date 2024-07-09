package main

import (
	"drone/pkg/protocol"
	"encoding/binary"
)

func (app *App) makeGyro(n int) *protocol.Message {
	app.mx.RLock()
	defer app.mx.RUnlock()

	dat := protocol.Make8b(app.data.roll, app.data.pitch, app.data.yaw, 10, app.data.flags, n)

	return protocol.NewMessage(0x8b, dat)
}

func (app *App) makeLL(n int) *protocol.Message {
	app.mx.RLock()
	defer app.mx.RUnlock()

	dat := protocol.Make8c(app.data.lat, app.data.lon, app.data.alt, app.data.vsp, app.data.hsp)
	return protocol.NewMessage(0x8c, dat)
}

func (app *App) makeBatt(n int) *protocol.Message {
	app.mx.RLock()
	defer app.mx.RUnlock()

	dat := make([]byte, 8)
	le := binary.LittleEndian

	dat[0] = 32 // em
	le.PutUint16(dat[1:], uint16(app.data.batt*100)<<2&0x3ffc)

	dat[7] = byte(n % 256) // 1 - 1 - optical, 0 - gps

	le.PutUint16(dat[5:], uint16(app.data.dist)<<1|1) // 5,6 - dist << 1

	return protocol.NewMessage(0x8f, dat)
}
