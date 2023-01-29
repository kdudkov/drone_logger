package protocol

type CommandBuilder struct {
	lr, fb, ud, rot, cmd, gimbal byte
	fast                         bool
}

func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{
		lr:   128, // 256 - right
		fb:   128, // 256 - forward
		ud:   128, // 256 - up
		rot:  128, // 256 - clockwise
		cmd:  0,
		fast: false,
	}
}

func (cb *CommandBuilder) withCmd(cmd byte) *CommandBuilder {
	cb.cmd = cmd
	return cb
}

func (cb *CommandBuilder) Up(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.ud = 128 + v
	return cb
}

func (cb *CommandBuilder) Down(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.ud = 128 - v
	return cb
}

func (cb *CommandBuilder) Right(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.lr = 128 + v
	return cb
}

func (cb *CommandBuilder) Left(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.lr = 128 - v
	return cb
}

func (cb *CommandBuilder) Forward(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.fb = 128 + v
	return cb
}

func (cb *CommandBuilder) Back(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.fb = 128 - v
	return cb
}

func (cb *CommandBuilder) Cw(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.rot = 128 + v
	return cb
}

func (cb *CommandBuilder) Ccw(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.rot = 128 - v
	return cb
}

func (cb *CommandBuilder) Build() []byte {
	res := make([]byte, 13)
	res[0] = cb.lr
	res[1] = cb.fb
	res[2] = cb.ud
	res[3] = cb.rot
	res[4] = 0x20
	if cb.fast {
		res[5] = 0x18
	} else {
		res[5] = 0x08
	}
	res[6] = cb.cmd
	res[7] = cb.gimbal
	res[8] = 1
	return createMessage(1, res)
}
