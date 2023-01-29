package protocol

import (
	"context"
	"io"
)

type Drone interface {
	Start(ctx context.Context)
	SetWriter(writer io.Writer)
	AddCommand(b []byte)
}
