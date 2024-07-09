package protocol

import (
	"context"
	"encoding/hex"
	"io"
)

type Drone interface {
	Start(ctx context.Context)
	Stop()
	SetLogger(logger io.Writer)
	AddCommand(m *Message)
}

func str2byte(s string) []byte {
	r, _ := hex.DecodeString(s)
	return r
}
