package csp

import (
	"errors"
	"fmt"
	"machine"
	"time"
)

const DEBUG = false

var ErrNoData = errors.New("No data available")
var ErrWrongChecksum = errors.New("Wrong checksum")
var ErrWrite = errors.New("Write failed")
var ErrWriteLength = errors.New("Write failed to send all bytes")

// States
const (
	STATE_IDLE byte = iota
	STATE_HEADER
	STATE_LENGTH
	STATE_COMMAND
	STATE_PAYLOAD
	STATE_CHECKSUM
)

type Adapter struct {
	uart *machine.UART

	state   byte
	Message Message
}

func NewAdapter(uart *machine.UART) *Adapter {
	return &Adapter{
		uart: uart,
	}
}

// Send a message.
func (a *Adapter) Send(m *Message) error {
	bytes := m.Bytes()
	if DEBUG {
		fmt.Printf("%s SEND ", time.Now().Format("15:04:05.000"))
		for _, b := range bytes {
			fmt.Printf(" %02X", b)
		}
		fmt.Println()
	}
	n, err := a.uart.Write(bytes)
	if err != nil {
		return ErrWrite
	}
	if n != len(bytes) {
		return ErrWriteLength
	}
	return nil
}

// Receive a message; returns nil if no message is available (yet).
func (a *Adapter) Receive() (*Message, error) {
	for {
		b, err := a.uart.ReadByte()
		if err != nil {
			return nil, ErrNoData
		}
		switch a.state {
		case STATE_IDLE:
			if b == '$' {
				if DEBUG {
					fmt.Printf("%s IDLE %02X\n", time.Now().Format("15:04:05.000"), b)
				}
				a.Message.Header[0] = b
				a.state = STATE_HEADER
			}
		case STATE_HEADER:
			if b == 'C' {
				if DEBUG {
					fmt.Printf("%s HEADER %02X\n", time.Now().Format("15:04:05.000"), b)
				}
				a.Message.Header[1] = b
				a.state = STATE_LENGTH
			} else {
				a.state = STATE_IDLE
			}
		case STATE_LENGTH:
			if DEBUG {
				fmt.Printf("%s LENGTH %02X\n", time.Now().Format("15:04:05.000"), b)
			}
			if b > MAX_PAYLOAD {
				a.state = STATE_IDLE
				continue
			}
			a.Message.Length = b
			a.Message.Payload = []byte{}
			a.Message.Checksum = b
			a.state = STATE_COMMAND
		case STATE_COMMAND:
			if DEBUG {
				fmt.Printf("%s COMMAND %02X\n", time.Now().Format("15:04:05.000"), b)
			}
			a.Message.Command = b
			a.Message.Checksum ^= b
			a.state = STATE_PAYLOAD
		case STATE_PAYLOAD:
			a.Message.Payload = append(a.Message.Payload, b)
			a.Message.Checksum ^= b
			if len(a.Message.Payload) == int(a.Message.Length) {
				a.state = STATE_CHECKSUM
			}
		case STATE_CHECKSUM:
			if DEBUG {
				fmt.Printf("%s PAYLOAD ", time.Now().Format("15:04:05.000"))
				for _, bb := range a.Message.Bytes() {
					fmt.Printf(" %02X", bb)
				}
				fmt.Println()
				fmt.Printf("%s CHECKSUM expected %02X ?= %02X actual\n", time.Now().Format("15:04:05.000"), a.Message.Checksum, b)
			}
			m := a.Message
			a.Message = Message{}
			a.state = STATE_IDLE
			if m.Checksum == b {
				return &m, nil
			} else {
				return nil, ErrWrongChecksum
			}
		}
	}
}

// Reset the state machine and clear the message buffer.
func (a *Adapter) Reset() {
	a.state = STATE_IDLE
	a.Message = Message{}
	for {
		_, err := a.uart.ReadByte()
		if err != nil {
			return
		}
	}
}
