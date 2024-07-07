package main

import (
	"errors"
	"fmt"
	"fpvc-gadget/src/csp"
	"time"
)

const (
	S_TEAM   = 72
	S_PLAYER = 73
)

var (
	ErrTimeout       = errors.New("Request Timeout")
	ErrWrongId       = errors.New("Wrong ID")
	ErrWrongVerson   = errors.New("Wrong Version")
	ErrWrongChecksum = errors.New("Wrong Checksum")
	ErrSetFailed     = errors.New("Set Failed")
)

var buf []byte = make([]byte, 1000)

type Settings struct {
	id        byte
	data      []byte // used for communication
	dirtyData []byte // used for changes
}

var settings = Settings{
	data:      make([]byte, 110),
	dirtyData: make([]byte, 110),
}

func (s *Settings) Fetch(id byte) error {
	fmt.Printf("Settings: Fetch %X\r\n", id)

	s.id = id

	request := csp.NewMessage(csp.COMMAND_CFG_GET_REQ, []byte{id, 0, 110})
	network.Reset()
	err := network.Send(request)
	if err != nil {
		println("Settings: Fetch Send error ", err.Error())
		return err
	}

	timeout := time.Now().Add(1 * time.Second)
	for time.Now().Before(timeout) {
		response, err := network.Receive()
		if err != nil && err != csp.ErrNoData {
			println("Settings: Fetch Receive error ", err.Error())
			return err
		}
		if response == nil {
			continue
		}
		if response.Command == csp.COMMAND_CFG_GET_RSP && response.Payload[0] == id {
			copy(s.data, response.Payload[2:]) // skip id and offset
			copy(s.dirtyData, response.Payload[2:])
			return nil
		}
	}

	return ErrTimeout
}

func (s *Settings) Commit(offset, length byte) error {
	println("Settings: Commit")
	needsPush := false
	for i := byte(0); i < length; i++ {
		if s.dirtyData[offset+i] != s.data[offset+i] {
			needsPush = true
			break
		}
	}
	if !needsPush {
		println("Settings: Commit no changes")
		return nil
	}
	dataToSend := make([]byte, 1+1+length)
	dataToSend[0] = s.id
	dataToSend[1] = offset
	copy(dataToSend[2:], s.dirtyData[offset:offset+length])

	request := csp.NewMessage(csp.COMMAND_CFG_SET_REQ, dataToSend)

	for attempts := 0; attempts < 1; attempts++ {
		network.Reset()
		err := network.Send(request)
		if err != nil {
			println("Settings: Commit send error ", err.Error())
			continue
		}
		timeout := time.Now().Add(1 * time.Second)
		var response *csp.Message
		for time.Now().Before(timeout) {
			response, err = network.Receive()
			if err != nil && err != csp.ErrNoData {
				println("Settings: Commit receive error ", err.Error())
			}
			if response != nil {
				break
			}
		}
		if response == nil {
			println("Settings: Commit receive timeout")
			continue
		}
		if response.Command == csp.COMMAND_CFG_SET_RSP {
			id := (s.dirtyData[S_TEAM] << 4) | s.dirtyData[S_PLAYER]
			if id != response.Payload[0] {
				return ErrSetFailed
			}
			for i := byte(1); i < length; i++ {
				if dataToSend[i] != response.Payload[i] {
					return ErrSetFailed
				}
			}
			s.id = id
			copy(s.data[offset:offset+length], s.dirtyData[offset:offset+length])
			return nil
		}
	}
	return ErrSetFailed
}

func (s *Settings) Get(address byte) byte {
	return s.dirtyData[address]
}

func (s *Settings) Set(address byte, value byte) {
	s.dirtyData[address] = value
}

func (s *Settings) GetBits(address byte, mask byte) byte {
	return (settings.dirtyData[address] & mask) >> s.maskOffset(mask)
}

func (s *Settings) SetBits(address byte, mask byte, value byte) {
	value <<= s.maskOffset(mask)
	settings.dirtyData[address] = (settings.dirtyData[address] & ^mask) | (value & mask)
}

func (s *Settings) maskOffset(mask byte) byte {
	offset := byte(0)
	for mask&1 == 0 {
		offset++
		mask >>= 1
	}
	return offset
}
