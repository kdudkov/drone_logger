package protocol

type CommandBuilder struct {
	roll, pitch, throttle, yaw, cmd, gimbal byte
	fast                                    bool
	remote                                  bool
}

func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{
		roll:     128, // 256 - right
		pitch:    128, // 256 - forward
		throttle: 128, // 256 - up
		yaw:      128, // 256 - clockwise
		gimbal:   0,
		cmd:      0,
		fast:     false,
		remote:   true,
	}
}

func (cb *CommandBuilder) WithCmd(cmd byte) *CommandBuilder {
	cb.cmd = cmd
	return cb
}

func (cb *CommandBuilder) Up(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.throttle = 128 + v
	cb.remote = false
	return cb
}

func (cb *CommandBuilder) Down(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.throttle = 128 - v
	cb.remote = false
	return cb
}

func (cb *CommandBuilder) Right(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.roll = 128 + v
	cb.remote = false
	return cb
}

func (cb *CommandBuilder) Left(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.roll = 128 - v
	cb.remote = false
	return cb
}

func (cb *CommandBuilder) Forward(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.pitch = 128 + v
	cb.remote = false
	return cb
}

func (cb *CommandBuilder) Back(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.pitch = 128 - v
	cb.remote = false
	return cb
}

func (cb *CommandBuilder) Cw(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.yaw = 128 + v
	cb.remote = false
	return cb
}

func (cb *CommandBuilder) Ccw(v byte) *CommandBuilder {
	if v > 127 {
		v = 127
	}
	cb.yaw = 128 - v
	cb.remote = false
	return cb
}

func (cb *CommandBuilder) GimbalUp() *CommandBuilder {
	cb.gimbal = 0x40
	return cb
}

func (cb *CommandBuilder) GimbalDown() *CommandBuilder {
	cb.gimbal = 0x80
	return cb
}

func (cb *CommandBuilder) Build() *Message {
	res := make([]byte, 13)
	res[0] = cb.roll
	res[1] = cb.pitch
	res[2] = cb.throttle
	res[3] = cb.yaw
	res[4] = 0x20
	if cb.fast {
		res[5] = 0x18
	} else {
		res[5] = 0x08
	}
	res[6] = cb.cmd
	res[7] = cb.gimbal

	if !cb.remote {
		res[8] = 1
	}

	return NewMessage(1, res)
}
