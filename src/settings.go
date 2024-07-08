package main

import (
	"errors"
	"fmt"
	"time"

	csp "github.com/ysoldak/fpvc-serial-protocol"
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

	// Send request
	network.Reset() // clear buffer expecting a response
	request := csp.NewConfigGetRequest(id, 0, 110)
	println("Settings: Fetch Send")
	err := network.Send(request.Message())
	if err != nil {
		println("Settings: Fetch Send error ", err.Error())
		return err
	}

	// Wait for response
	message, err := network.Wait(csp.CommandConfigGet, csp.DirResponse, 1*time.Second)
	if err != nil {
		return ErrTimeout
	}
	configGetResponse := csp.NewConfigGetResponseFromMessage(message)
	if configGetResponse.ID != id {
		return ErrWrongId
	}
	copy(s.data, configGetResponse.Data)
	copy(s.dirtyData, configGetResponse.Data)
	return nil
}

func (s *Settings) Commit(offset, length byte) error {
	println("Settings: Commit")

	// Check if there are any changes
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

	// Send request
	network.Reset() // clear buffer expecting a response
	request := csp.NewConfigSetRequest(s.id, offset, s.dirtyData[offset:offset+length])
	err := network.Send(request.Message())
	if err != nil {
		println("Settings: Commit send error ", err.Error())
		return ErrSetFailed
	}

	// Wait for response
	expectedID := (s.dirtyData[S_TEAM] << 4) | s.dirtyData[S_PLAYER] // ID may change if team or player number changes
	message, err := network.Wait(csp.CommandConfigSet, csp.DirResponse, 1*time.Second)
	if err != nil {
		return ErrSetFailed
	}
	configSetResponse := csp.NewConfigSetResponseFromMessage(message)
	if configSetResponse.ID != expectedID {
		return ErrSetFailed
	}
	// check if data was set correctly
	for i := byte(0); i < length; i++ {
		if s.dirtyData[offset+i] != configSetResponse.Data[i] {
			return ErrSetFailed
		}
	}
	// apply changes and return
	s.id = expectedID
	copy(s.data[offset:offset+length], s.dirtyData[offset:offset+length])
	return nil
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
