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

	request := csp.NewMessage(csp.COMMAND_CFG_GET_REQ, []byte{id})
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
			copy(s.data, response.Payload[1:])
			copy(s.dirtyData, response.Payload[1:])
			return nil
		}
	}

	return ErrTimeout
}

func (s *Settings) Commit() error {
	println("Settings: commit")
	needsPush := false
	for i := 0; i < 110; i++ {
		if s.dirtyData[i] != s.data[i] {
			needsPush = true
			break
		}
	}
	if !needsPush {
		return nil
	}
	dataToSend := make([]byte, 1+110)
	dataToSend[0] = s.id
	copy(dataToSend[1:], s.dirtyData)

	request := csp.NewMessage(csp.COMMAND_CFG_SET_REQ, dataToSend)

	for attempts := 0; attempts < 1; attempts++ {
		network.Reset()
		err := network.Send(request)
		if err != nil {
			println("Settings: Commit Send error ", err.Error())
			continue
		}
		timeout := time.Now().Add(1 * time.Second)
		var response *csp.Message
		for time.Now().Before(timeout) {
			response, err = network.Receive()
			if err != nil && err != csp.ErrNoData {
				println("Settings: Commit Receive error ", err.Error())
			}
			if response == nil {
				continue
			}
		}
		if response == nil {
			println("Settings: commit receive timeout")
			continue
		}
		if response.Command == csp.COMMAND_CFG_SET_RSP {
			id := (s.dirtyData[S_TEAM] << 4) + s.dirtyData[S_PLAYER]
			if id != response.Payload[0] {
				return ErrSetFailed
			}
			for i := 1; i < 111; i++ {
				if dataToSend[i] != response.Payload[i] {
					return ErrSetFailed
				}
			}
			s.id = id
			copy(s.data, s.dirtyData)
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
