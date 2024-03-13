package main

import (
	"errors"
	"fmt"
	"time"
)

const SettingsVersion = 2

const (
	S_LIFE = 70
	S_AMMO = 71
	S_TEAM = 72
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
	version   byte
	data      []byte // used for communication
	dirtyData []byte // used for changes
}

var settings = Settings{
	version:   SettingsVersion,
	data:      make([]byte, 110),
	dirtyData: make([]byte, 110),
}

func (s *Settings) Fetch(id byte) error {
	fmt.Printf("Settings: fetch %X\r\n", id)

	s.id = id
	for serial.uart.Buffered() > 0 {
		serial.uart.Read(buf)
	}
	serial.uart.WriteByte(0x72)
	serial.uart.WriteByte(id)

	time.Sleep(1 * time.Second)

	err := s.Read()
	if err != nil {
		return err
	}

	copy(s.data, buf[2:112])
	copy(s.dirtyData, buf[2:112])

	return nil
}

func (s *Settings) Read() error {
	println("Settings: read")

	i := 0
	for serial.uart.Buffered() < 113 {
		if i > 50 {
			return ErrTimeout
		}
		time.Sleep(100 * time.Millisecond)
		i++
	}

	_, err := serial.uart.Read(buf)
	if err != nil {
		return err
	}

	// filter out noise
	for len(buf) > 0 && buf[0] == 0xFF {
		buf = buf[1:]
	}

	if buf[0] != s.id {
		return ErrWrongId
	}

	if buf[1] != s.version {
		return ErrWrongVerson
	}

	checksum := byte(0)
	for i := 0; i < 112; i++ {
		checksum += buf[i]
	}
	if buf[112] != checksum {
		return ErrWrongChecksum
	}

	return nil
}

func (s *Settings) Push() error {
	println("Settings: push")

	display.Print(120, 60, "*")
	display.Show()

	// send new config
	serial.uart.WriteByte(0x74)
	serial.uart.WriteByte(s.id)
	serial.uart.WriteByte(s.version)
	checksum := s.id + s.version
	for i := 0; i < 110; i++ {
		serial.uart.WriteByte(s.data[i])
		checksum += s.data[i]
	}
	serial.uart.WriteByte(checksum)

	// team may have been changed
	s.id = (s.data[S_TEAM] << 4) + s.id&0x0F

	// wait confirmation
	err := s.Read()
	if err != nil {
		display.Erase(120, 60, "*")
		display.Print(120, 60, "?") // network error
		display.Show()
		return err
	}

	for i := 0; i < 110; i++ {
		if buf[i+2] != s.data[i] {
			display.Erase(120, 60, "*")
			display.Print(120, 60, "!") // data mismatch, probably failed to set or rejected
			display.Show()
			return ErrSetFailed
		}
	}

	display.Erase(120, 60, "*")
	display.Show()
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

func (s *Settings) Commit() error {
	println("Settings: commit")
	needsPush := false
	for i := 0; i < 110; i++ {
		if s.dirtyData[i] != s.data[i] {
			needsPush = true
			break
		}
	}
	if needsPush {
		copy(s.data, s.dirtyData)
		return s.Push()
	}
	return nil
}
