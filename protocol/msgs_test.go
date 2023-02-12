package protocol

import (
	"testing"
)

var msgs = []string{
	"588a0c584c30313253465878800c0810",
	"588b0e3900f3007e3dfb00290000000007d9",
	"588c0d0000000000000000e8030000006a",
	"588f08c0f00c0000010000ba",
	"58fe1948462d584c2d584c303132533031312e3032312e31323232058c",
}

func TestCreateMessage(t *testing.T) {
	msg := createMessage(0x0b, []byte{0, 0, 0, 0, 0, 0, 0, 0})

	if len(msg) != 12 {
		t.Fail()
	}

	expected := []byte{0x68, 0xb, 8, 0, 0, 0, 0, 0, 0, 0, 0, 3}

	for i, c := range msg {
		if expected[i] != c {
			t.Fail()
		}
	}

}
