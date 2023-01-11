package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func sender4(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	for ctx.Err() == nil {
		select {
		case <-ticker.C:
			conn, err := net.Dial("udp", "192.168.0.1:40000")
			if err != nil {
				panic(err)
			}
			err = conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
			if err != nil {
				panic(err)
			}

			_, err = conn.Write([]byte{0x63, 0x63, 1, 0, 0, 0, 0})
			if err != nil {
				panic(err)
			}
			_ = conn.Close()
		case <-ctx.Done():
			break
		}
	}
	ticker.Stop()
}

func sender5(ctx context.Context, conn net.Conn, ch chan []byte) {
	for {
		select {
		case data := <-ch:
			if _, err := conn.Write(data); err != nil {
				fmt.Println(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func pinger(ctx context.Context, ch chan []byte) {
	ticker := time.NewTicker(time.Second * 3)
	for ctx.Err() == nil {
		select {
		case <-ticker.C:
			ch <- createMessage(0x0b, []byte{0, 0, 0, 0, 0, 0, 0, 0})
		case <-ctx.Done():
			break
		}
	}
	ticker.Stop()
}

func reader(ctx context.Context, conn net.Conn) {
	buf := make([]byte, 4096)
	for ctx.Err() == nil {
		n, err := conn.Read(buf)
		if err != nil {
			continue
		}
		msg := make([]byte, n)
		copy(msg, buf[:n])
		if err := printMessage(msg); err != nil {
			fmt.Println(err)
		}
	}
}

func printMessage(msg []byte) error {
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

	switch msg[1] {
	case 0x8b:
		//roll := float64(int16(binary.LittleEndian.Uint16(msg[3:5]))) / 100
		//pitch := float64(int16(binary.LittleEndian.Uint16(msg[5:7]))) / 100
		//yaw := float64(int16(binary.LittleEndian.Uint16(msg[7:9]))) / 100
		//fmt.Printf("roll: %.2f pitch %.2f yaw %.2f\n", roll, pitch, yaw)
		//for _, x := range msg[9 : len(msg)-1] {
		//	fmt.Printf("%.2x ", x)
		//}
		//fmt.Println()
	case 0x8c:
		lon := float64(int32(binary.LittleEndian.Uint32(msg[3:7]))) / 10000000
		lat := float64(int32(binary.LittleEndian.Uint32(msg[7:11]))) / 10000000
		fmt.Printf("lat %.5f lon %.5f\n", lat, lon)
		for _, x := range msg[11 : len(msg)-1] {
			fmt.Printf("%.2x ", x)
		}
		fmt.Println()
	case 0x8f:
		em := msg[3]
		fmt.Printf("em: %d\n", em)
		for _, x := range msg[4 : len(msg)-1] {
			fmt.Printf("%.2x ", x)
		}
		fmt.Println()
	case 0xfe:
		s := string(msg[3 : len(msg)-2])
		fmt.Printf("ver: %s\n", s)
	default:
		fmt.Printf("new message %.2x\n", msg[1])
		for _, x := range msg {
			fmt.Printf("%.2x ", x)
		}
		fmt.Println()
	}

	return nil
}

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

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := net.Dial("udp", "192.168.0.1:50000")
	if err != nil {
		panic(err)
	}

	ch := make(chan []byte, 10)

	//go sender4(ctx)
	go sender5(ctx, conn, ch)

	ch <- createMessage(0x0b, []byte{0, 0, 0, 0, 0, 0, 0, 0})
	go pinger(ctx, ch)
	go reader(ctx, conn)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-c
	cancel()
}
