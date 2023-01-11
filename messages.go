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

func createRcCommand(lr, fb, th, turn, command byte) []byte {
	res := make([]byte, 8)
	res[0] = 0x66
	res[1] = lr
	res[2] = fb
	res[3] = th
	res[4] = turn
	res[5] = command
	res[6] = lr ^ fb ^ th ^ turn
	res[7] = 0x99

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

	app.data.Store(msg[1], msg[3:len(msg)-2])

	switch msg[1] {
	case 0x8b:
		app.info.put("roll", float64(int16(binary.LittleEndian.Uint16(msg[3:5])))/100)
		app.info.put("pitch", float64(int16(binary.LittleEndian.Uint16(msg[5:7])))/100)
		app.info.put("yaw", float64(int16(binary.LittleEndian.Uint16(msg[7:9])))/100)
	case 0x8c:
		app.info.put("lon", float64(int32(binary.LittleEndian.Uint32(msg[3:7])))/10000000)
		app.info.put("lat", float64(int32(binary.LittleEndian.Uint32(msg[7:11])))/10000000)
	case 0x8f:
		app.info.put("em", msg[3])
		app.info.put("battery", float64(int16(binary.LittleEndian.Uint16(msg[4:6])))/400)
	case 0xfe:
		app.info.put("ver", string(msg[3:len(msg)-2]))
	}

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
