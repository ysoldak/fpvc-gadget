package csp

import (
	"errors"
	"machine"
)

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
	STATE_DATA
	STATE_CHECKSUM
)

type Adapter struct {
	uart *machine.UART

	state   byte
	message Message
}

func NewAdapter(uart *machine.UART) *Adapter {
	return &Adapter{
		uart: uart,
	}
}

// Send a message.
func (csp *Adapter) Send(m *Message) error {
	bytes := m.Bytes()
	n, err := csp.uart.Write(bytes)
	if err != nil {
		println(err.Error())
		return ErrWrite
	}
	if n != len(bytes) {
		return ErrWriteLength
	}
	return nil
}

// Receive a message; returns nil if no message is available (yet).
func (csp *Adapter) Receive() (*Message, error) {
	for {
		b, err := csp.uart.ReadByte()
		if err != nil {
			return nil, ErrNoData
		}
		switch csp.state {
		case STATE_IDLE:
			if b == '$' {
				csp.message.Header[0] = b
				csp.state = STATE_HEADER
			}
		case STATE_HEADER:
			if b == 'C' {
				csp.message.Header[1] = b
				csp.state = STATE_LENGTH
			} else {
				csp.state = STATE_IDLE
			}
		case STATE_LENGTH:
			csp.message.Length = b
			csp.message.Checksum = b
			csp.state = STATE_COMMAND
		case STATE_COMMAND:
			csp.message.Command = b
			csp.message.Checksum ^= b
			csp.state = STATE_DATA
		case STATE_DATA:
			csp.message.Data = append(csp.message.Data, b)
			csp.message.Checksum ^= b
			if len(csp.message.Data) == int(csp.message.Length) {
				csp.state = STATE_CHECKSUM
			}
		case STATE_CHECKSUM:
			m := csp.message
			csp.message = Message{}
			csp.state = STATE_IDLE
			if csp.message.Checksum == b {
				return &m, nil
			} else {
				return nil, ErrWrongChecksum
			}
		}
	}
}

// Reset the state machine and clear the message buffer.
func (csp *Adapter) Reset() {
	csp.state = STATE_IDLE
	csp.message = Message{}
	for csp.uart.Buffered() > 0 {
		_, err := csp.uart.ReadByte()
		if err != nil {
			return
		}
	}
}
