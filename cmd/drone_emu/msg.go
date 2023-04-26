package main

import (
	"encoding/binary"
)

func (app *App) makeGyro() []byte {
	app.mx.RLock()
	defer app.mx.RUnlock()
	// 14: 3900 f300 7e3d fb00 290000 00 00 07

	dat := make([]byte, 14)
	le := binary.LittleEndian
	le.PutUint16(dat, uint16(app.data.roll*100))
	le.PutUint16(dat[2:], uint16(app.data.pitch*100))
	le.PutUint16(dat[4:], uint16(app.data.yaw*100))

	dat[6] = app.n
	dat[7] = app.n
	dat[8] = 0 // ? - lo-power alarm

	dat[9] = 0 //  4 - locked  5 - inAir/ground 6 - returning 7 - right icons 8 - in route
	if !app.data.locked {
		dat[9] |= 1 << 3
	}
	if app.data.inAir {
		dat[9] |= 1 << 4
	}
	if app.data.route {
		dat[9] |= 1 << 7
	}

	dat[11] = app.data.sats << 2 //  bit 0 - landing 7 - gps good 2..6 - gps num
	if app.data.sats > 5 {
		dat[11] |= 1 << 7
	}

	if app.data.landing {
		dat[11] |= 1
	}

	dat[10] = app.n
	dat[12] = app.n

	dat[13] = 1 //  2 - 50/100%  3 - remote off
	if app.data.hiSpeed {
		dat[13] |= 2
	}

	//fmt.Println(app.n)
	app.n++

	return createMessage(0x8b, dat)
}

func (app *App) makeLL() []byte {
	app.mx.RLock()
	defer app.mx.RUnlock()
	// 13: 00000000 00000000 e803 0000 00

	dat := make([]byte, 13)
	le := binary.LittleEndian
	le.PutUint32(dat, uint32(app.data.lon*10000000))
	le.PutUint32(dat[4:], uint32(app.data.lat*10000000))
	le.PutUint16(dat[8:], uint16(app.data.alt*10+1000)) // alt
	le.PutUint16(dat[10:], uint16(app.data.dist/1.6))   // dist

	dat[12] = uint8(-app.data.vsp * 10) //vs

	return createMessage(0x8c, dat)
}

func (app *App) makeBatt() []byte {
	app.mx.RLock()
	defer app.mx.RUnlock()
	//	case 0x8f: // 8: c0 f00c 0000010000
	//                   xx xxxx ..xxxxxx..
	//		s.dataAdder("em", data[0])
	//		s.dataAdder("battery", float64((le.Uint16(data[1:3])&0x3ffc)>>2)/100)
	//		s.dataAdder("f1", data[1]&0xc0>>6)
	//		s.dataAdder("f2", data[1]&3)

	dat := make([]byte, 8)
	le := binary.LittleEndian

	dat[0] = 22 // em
	le.PutUint16(dat[1:], uint16(870)<<2)
	dat[4] = 0                      // 1 - 1 - optical, 0 - gps
	le.PutUint16(dat[5:], 777<<1|1) // 5,6 - dist << 1

	return createMessage(0x8f, dat)
}
