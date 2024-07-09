package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type Message struct {
	Back bool
	Code byte
	Data []byte
}

func NewMessage(code byte, dat []byte) *Message {
	return &Message{Code: code, Data: dat}
}

func (m *Message) Empty() bool {
	return m == nil || len(m.Data) == 0
}

func (m *Message) String() string {
	return fmt.Sprintf("code: %.2x, data: %s", m.Code, hex.EncodeToString(m.Data))
}

func (m *Message) Marshal() []byte {
	res := make([]byte, len(m.Data)+4)

	if m.Back {
		res[0] = 0x58 // drone to gs
	} else {
		res[0] = 0x68 // gs to drone
	}
	res[1] = m.Code
	res[2] = byte(len(m.Data))
	for i, c := range m.Data {
		res[i+3] = c
	}

	var csum byte
	for _, c := range res[1 : len(res)-1] {
		csum ^= c
	}

	res[len(res)-1] = csum

	return res
}

func (m *Message) Unmarshal(d []byte) error {
	if d[0] != 0x68 && d[0] != 0x58 {
		return fmt.Errorf("invalid first byte")
	}

	if d[0] != 0x58 {
		m.Back = true
	}

	if len(d) < 3 || len(d)-4 != int(d[2]) {
		return fmt.Errorf("invalid lenght")
	}

	var csum byte
	for _, c := range d[1 : len(d)-1] {
		csum ^= c
	}

	if csum != d[len(d)-1] {
		return fmt.Errorf("invalid checksum: %.2x %.2x", csum, d[len(d)-1])
	}

	m.Code = d[1]
	m.Data = make([]byte, len(d)-4)
	copy(m.Data, d[3:len(d)-2])

	return nil
}

func Make8b(roll, pitch, yaw float64, sats byte, flags map[string]bool, n int) []byte {
	dat := make([]byte, 14)
	le := binary.LittleEndian
	le.PutUint16(dat, uint16(roll*100))
	le.PutUint16(dat[2:], uint16(pitch*100))
	le.PutUint16(dat[4:], uint16(yaw*100))

	dat[8] = 0 // 128, 64 - lo-power alarm

	dat[9] = 0 //  1,2,4 - none, 8 - armed  16 - inAir/ground 32 - returning 64 - right icons 128 - in route

	if flags["armed"] {
		dat[9] |= 8
	}
	if flags["in_air"] {
		dat[9] |= 16
	}
	if flags["returning"] {
		dat[9] |= 32
	}
	if flags["route"] {
		dat[9] |= 128
	}

	dat[11] = sats << 2 //  1 - landing, 2..7 - sat num, 128 - gps good

	if flags["landing"] {
		dat[11] |= 1
	}

	if flags["sat_good"] {
		dat[11] |= 128
	}

	dat[13] = 0 //  2 - 50/100%  ? - remote off

	if flags["hi_speed"] {
		dat[13] |= 2
	}

	return dat
}

func parse8b(data []byte, adder func(string, any)) {
	le := binary.LittleEndian

	adder("roll", float64(int16(le.Uint16(data[0:2])))/100)
	adder("pitch", float64(int16(le.Uint16(data[2:4])))/100)
	adder("yaw", float64(int16(le.Uint16(data[4:6])))/100)

	adder("armed", data[9]&8 > 0)

	inAir := data[9]&16 > 0
	adder("in_air", inAir)

	if !inAir {
		adder("state", "ground")
	} else {
		adder("state", "air")
	}

	if data[9]&32 > 0 {
		adder("state", "returning")
	}

	if data[9]&128 > 0 {
		adder("state", "route")
	}

	adder("sat", int((data[11]&0x7c)>>2))
	adder("sat_good", data[11]&128 > 0)

	if data[11]&1 > 0 {
		adder("state", "landing")
	}

	if len(data) > 13 {
		adder("hi_speed", data[13]&2 > 0)
		adder("remote", data[13]&4 == 0)
	}
}

func parse8c(data []byte, adder func(string, any)) {
	le := binary.LittleEndian

	adder("lon", float64(int32(le.Uint32(data[0:4])))/10000000)
	adder("lat", float64(int32(le.Uint32(data[4:8])))/10000000)
	adder("alt", float64(int16(le.Uint16(data[8:10])))/10-100)
	//adder("dist", float64(int16(le.Uint16(data[10:12])))*1.6)

	if len(data) > 12 {
		adder("hsp", -float64(int8(data[12]))/10)
	}
}

func Make8c(lat, lon, alt, vsp, hsp float64) []byte {
	dat := make([]byte, 14)
	le := binary.LittleEndian
	le.PutUint32(dat, uint32(lon*10000000))
	le.PutUint32(dat[4:], uint32(lat*10000000))
	le.PutUint16(dat[8:], uint16(alt*10+1000))

	dat[13] = uint8(-vsp * 10)
	dat[12] = uint8(-hsp * 10)

	//dat[13] = 55

	return dat
}

func parse8f(data []byte, adder func(string, any)) {
	le := binary.LittleEndian

	adder("em", int(data[0])+int((data[1]&0xc0)>>6)*256)
	adder("battery", float64((le.Uint16(data[1:3])&0x3ffc)>>2)/100)
	adder("optical", data[4]&1 > 0)
	adder("dist", le.Uint16(data[5:7])>>1)
}
