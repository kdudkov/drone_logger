package main

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func createMessage(code byte, data []byte) []byte {
	res := make([]byte, len(data)+4)
	res[0] = 0x68
	res[1] = code
	res[2] = byte(len(data))
	for i, c := range data {
		res[i+3] = c
	}
	var csum byte
	for _, c := range res[1 : len(res)-1] {
		csum ^= c
	}
	res[len(res)-1] = csum
	return res
}

func (app *App) processMessage(msg []byte) error {
	if len(msg) < 3 || len(msg)-4 != int(msg[2]) {
		return fmt.Errorf("invalid lenght")
	}

	var csum byte
	for _, c := range msg[1 : len(msg)-1] {
		csum ^= c
	}

	if csum != msg[len(msg)-1] {
		return fmt.Errorf("invalid checksum: %.2x %.2x", csum, msg[len(msg)-1])
	}

	data := msg[3 : len(msg)-2]
	app.data.Store(msg[1], data)

	le := binary.LittleEndian

	switch msg[1] {
	case 0x8b:
		app.info.put("roll", float64(int16(le.Uint16(data[0:2])))/100)
		app.info.put("pitch", float64(int16(le.Uint16(data[2:4])))/100)
		app.info.put("yaw", float64(int16(le.Uint16(data[4:6])))/100)
		app.info.put("sat", (data[11]&0x7c)>>2)
		app.info.put("s1", (data[11]&0x80)>>7)
		app.info.put("s2", data[11]&3)
	case 0x8c:
		app.info.put("lon", float64(int32(le.Uint32(data[0:4])))/10000000)
		app.info.put("lat", float64(int32(le.Uint32(data[4:8])))/10000000)
		app.info.put("hs", float64(int8(data[10]))/10)
	case 0x8f:
		app.info.put("em", data[0])
		app.info.put("battery", float64((le.Uint16(data[1:3])&0x3ffc)>>2)/100)
		app.info.put("f1", data[1]&0xc0>>6)
		app.info.put("f2", data[1]&3)
	case 0xfe:
		app.info.put("ver", string(data[1:len(data)-1]))
	}
	// c8 - 18 cc - 19 c4 - 17

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
