package protocol

import (
	"context"
	"encoding/hex"
	"io"
)

type Drone interface {
	Start(ctx context.Context)
	SetWriter(writer io.Writer)
	AddCommand(b []byte)
}

func str2byte(s string) []byte {
	r, _ := hex.DecodeString(s)
	return r
}
