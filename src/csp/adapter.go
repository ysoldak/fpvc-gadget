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
func (csp *Adapter) Send(m *Message) error {
	bytes := m.Bytes()
	// fmt.Printf("%s SEND ", time.Now().Format("15:04:05.000"))
	// for _, b := range bytes {
	// 	fmt.Printf(" %02X", b)
	// }
	// println()

	n, err := csp.uart.Write(bytes)
	if err != nil {
		// println(err.Error())
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
				// fmt.Printf("%s IDLE %02X\n", time.Now().Format("15:04:05.000"), b)
				csp.Message.Header[0] = b
				csp.state = STATE_HEADER
			}
		case STATE_HEADER:
			if b == 'C' {
				// fmt.Printf("%s HEADER %02X\n", time.Now().Format("15:04:05.000"), b)
				csp.Message.Header[1] = b
				csp.state = STATE_LENGTH
			} else {
				csp.state = STATE_IDLE
			}
		case STATE_LENGTH:
			// fmt.Printf("%s LENGTH %02X\n", time.Now().Format("15:04:05.000"), b)
			if b > MAX_PAYLOAD {
				csp.state = STATE_IDLE
				continue
			}
			csp.Message.Length = b
			csp.Message.Payload = []byte{}
			csp.Message.Checksum = b
			csp.state = STATE_COMMAND
		case STATE_COMMAND:
			// fmt.Printf("%s COMMAND %02X <\n", time.Now().Format("15:04:05.000"), b)
			csp.Message.Command = b
			csp.Message.Checksum ^= b
			csp.state = STATE_PAYLOAD
		case STATE_PAYLOAD:
			csp.Message.Payload = append(csp.Message.Payload, b)
			csp.Message.Checksum ^= b
			if len(csp.Message.Payload) == int(csp.Message.Length) {
				csp.state = STATE_CHECKSUM
			}
		case STATE_CHECKSUM:
			// fmt.Printf("%s CHECKSUM %02X vs %02X <\n", time.Now().Format("15:04:05.000"), b, csp.Message.Checksum)
			m := csp.Message
			csp.Message = Message{}
			csp.state = STATE_IDLE
			if m.Checksum == b {
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
	csp.Message = Message{}
	for {
		_, err := csp.uart.ReadByte()
		if err != nil {
			// println()
			return
		}
		// print(".")
	}
}
