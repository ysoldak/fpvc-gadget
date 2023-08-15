package main

import (
	"fmt"
	"time"
)

type Scan struct {
	devices []*Device

	cursor     int
	cursorPrev int

	redraw bool

	active bool
}

func (s *Scan) Open() {

	encoder.SetChangeHandler(nil)
	encoder.SetClickHandler(nil)

	s.active = true
	s.redraw = true
	s.cursor = 0
	s.cursorPrev = -1

	s.Show()

	encoder.SetChangeHandler(s.HandleChange)
	encoder.SetClickHandler(s.HandleClick)

	i := 0
	for {

		if i%500 == 0 { // every 5 sec
			serial.uart.WriteByte(0x70)
			display.Print(120, 60, "*")
			display.device.Display()
		}
		if (i-100)%500 == 0 { // every 5+1 sec
			display.Clear(120, 60, "*")
			display.device.Display()
		}

		s.Receive()

		s.Show()

		if i%100 == 0 {
			println(i)
		}
		time.Sleep(10 * time.Millisecond)

		i++

		if !s.active {
			device := s.devices[s.cursor]
			device.Open()

			encoder.SetChangeHandler(nil)
			encoder.SetClickHandler(nil)

			i = 0
			s.active = true
			s.redraw = true
			s.cursor = 0
			s.cursorPrev = -1
			s.devices = []*Device{}

			s.Show()

			encoder.SetChangeHandler(s.HandleChange)
			encoder.SetClickHandler(s.HandleClick)
		}
	}

}

func (s *Scan) Show() {
	if s.redraw {
		display.device.ClearDisplay()
		for i, dev := range s.devices {
			// fmt.Printf("%X | '%s'\r\n", dev.id, string(dev.name))
			display.Print(10, 10+int16(i)*10, fmt.Sprintf("  %X | %s ", dev.id, dev.name))
		}
		display.Print(10, 10+int16(s.cursor)*10, ">")
		s.redraw = false
		display.device.Display()
	}
	if s.cursorPrev != s.cursor {
		display.Clear(10, 10+int16(s.cursorPrev)*10, ">")
		display.Print(10, 10+int16(s.cursor)*10, ">")
		s.cursorPrev = s.cursor
		display.device.Display()
	}
}

func (s *Scan) Receive() {
	if serial.uart.Buffered() < 22 {
		return
	}
	key := byte(0)
	name := make([]byte, 10)
	for i := 0; i < 22; i++ {
		b, err := serial.uart.ReadByte()
		if err != nil {
			return
		}
		if i == 0 && 0xA0 < b && b < 0xFF {
			key = b
		}
		if 1 < i && i < 12 {
			if b == 0x2F {
				b = 0x20
			}
			name[i-2] = b
		}
	}
	found := false
	for i := range s.devices {
		if s.devices[i].id == key {
			s.devices[i].name = string(name)
			found = true
		}
	}
	if !found {
		device := NewDevice()
		device.id = key
		device.name = string(name)
		s.devices = append(s.devices, device)
		s.redraw = true
	}
}

func (s *Scan) HandleClick() {
	s.active = false
}

func (s *Scan) HandleChange(value int) int {
	s.cursorPrev = s.cursor
	s.cursor = value
	if value < 0 {
		s.cursor = 0
	}
	if value > len(s.devices)-1 {
		s.cursor = len(s.devices) - 1
	}
	return s.cursor
}
