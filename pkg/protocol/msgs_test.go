package protocol

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var msgs = []string{
	"588a0c584c30313253465878800c0810",
	"588b0e3900f3007e3dfb00290000000007d9",
	"588c0d0000000000000000e8030000006a",
	"588f08c0f00c0000010000ba",
	"58fe1948462d584c2d584c303132533031312e3032312e31323232058c",
}

func TestParse8b(t *testing.T) {
	m := make(map[string]any)

	adder := func(k string, v any) {
		m[k] = v
	}

	d, _ := hex.DecodeString("3900f3007e3dfb00290000000007")
	parse8b(d, adder)

	assert.Equal(t, true, m["hi_speed"])
	assert.Equal(t, true, m["locked"])
	assert.Equal(t, false, m["in_air"])
	assert.Equal(t, false, m["remote"])
	assert.Equal(t, "ground", m["state"])
	assert.Equal(t, false, m["sat_good"])
	assert.Equal(t, 0, m["sat"])

	assert.Equal(t, 0.57, m["roll"])
	assert.Equal(t, 157.42, m["yaw"])
	assert.Equal(t, 2.43, m["pitch"])
}

func TestParse8f(t *testing.T) {
	m := make(map[string]any)

	adder := func(k string, v any) {
		m[k] = v
	}

	d, _ := hex.DecodeString("08c0f00c0000010000")
	parse8f(d, adder)

	fmt.Println(m)
}

func Test8c(t *testing.T) {
	m := make(map[string]any)

	adder := func(k string, v any) {
		m[k] = v
	}

	lat, lon := 60.1, 30.2

	//dist := 120.0
	alt := 70.0
	vsp := 2.3
	hsp := 8.3

	d := Make8c(lat, lon, alt, vsp, hsp)
	parse8c(d, adder)

	assert.Equal(t, lat, m["lat"])
	assert.Equal(t, lon, m["lon"])
	assert.InDelta(t, alt, m["alt"], 0.1)

	if hsp == 0 {
		assert.Nil(t, m["hsp"])
	} else {
		assert.InDelta(t, hsp, m["hsp"], 0.1)
	}

	fmt.Println(m)
}
