package csp

const MAX_PAYLOAD = 111

// Commands
const (
	COMMAND_BEACON byte = 0x71 // ID[1], Name[10], Description[20]

	COMMAND_CFG_GET_REQ byte = 0x72 // ID[1]
	COMMAND_CFG_GET_RSP byte = 0x73 // ID[1], Data[110]
	COMMAND_CFG_SET_REQ byte = 0x74 // ID[1], Data[110]
	COMMAND_CFG_SET_RSP byte = 0x75 // ID[1], Data[110]

	COMMAND_HIT   byte = 0x82 // ID[1], Lives[1]
	COMMAND_CLAIM byte = 0x84 // ID[1], Power[1]
)

type Message struct {
	Header   [2]byte // '$' + 'C'
	Length   byte    // Length of the payload
	Command  byte    // 0x82 = Claim, 0x83 = Hit, etc
	Payload  []byte  // Data
	Checksum byte    // XOR of all bytes from length to the end of payload
}

func NewMessage(command byte, data []byte) *Message {
	checksum := byte(len(data)) ^ command
	for _, b := range data {
		checksum ^= b
	}
	return &Message{
		Header:   [2]byte{'$', 'C'},
		Length:   byte(len(data)),
		Command:  command,
		Payload:  data,
		Checksum: checksum,
	}
}

func (m *Message) Bytes() []byte {
	b := make([]byte, 4+len(m.Payload)+1)
	b[0] = m.Header[0]
	b[1] = m.Header[1]
	b[2] = m.Length
	b[3] = m.Command
	copy(b[4:], m.Payload)
	b[len(b)-1] = m.Checksum
	return b
}
