package main

import (
	"fmt"
	"time"
)

type SettingKind byte

const (
	SettingKindByte SettingKind = iota
	// SettingKindString
)

const (
	SettingShowDec byte = iota
	SettingShowHex
	SettingShowChar
)

type Setting struct {
	address byte
	value   []byte

	min     byte   // min value of each byte
	max     byte   // max value of each byte
	invalid []byte // if any of values are invalid and shall be hopped over

	cursor byte
	len    byte // total length of the value slice

	// visualisation
	kind   SettingKind
	show   byte
	title  string
	update bool
	active bool

	positionOffset int
}

func (s *Setting) Open(position int) {

	s.active = true
	s.cursor = 0

	encoder.SetClickHandler(s.HandleClick)
	encoder.SetChangeHandler(s.HandleChange)
	encoder.device.SetValue(int(s.value[s.cursor]))

	for s.active {
		if s.update {
			s.update = false
			s.Show(position)
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func (s *Setting) Show(position int) {
	display.Fill(20, int16(s.positionOffset+10*(position-1)+3), 128, 10, BLACK)
	switch s.show {
	case SettingShowDec:
		display.Print(1, int16(s.positionOffset+10*position), fmt.Sprintf("  %s: %d", s.title, s.value[0])) // TODO support global scroll
	case SettingShowHex:
		display.Print(1, int16(s.positionOffset+10*position), fmt.Sprintf("  %s: %X", s.title, s.value[0])) // TODO support global scroll
	case SettingShowChar:
		display.Print(1, int16(s.positionOffset+10*position), fmt.Sprintf("  %s: %s", s.title, string(s.value))) // TODO support global scroll
	}
	display.device.Display()
}

func (s *Setting) HandleClick() {
	s.cursor++
	if s.cursor >= s.len {
		s.active = false
		return
	}
	encoder.device.SetValue(int(s.value[s.cursor]))
}

func (s *Setting) HandleChange(value int) int {
	if s.kind == SettingKindByte {
		oldValue := s.value[s.cursor]
		newValue := byte(value)

		// filter invalid values by hopping over them in the same direction
		good := false
		for !good {
			good = true
			for _, v := range s.invalid {
				if newValue == v {
					good = false
					if newValue > oldValue {
						newValue++
					} else {
						newValue--
					}
					if newValue < s.min {
						newValue = s.max
					}
					if newValue > s.max {
						newValue = s.min
					}
					break
				}
			}
		}

		if newValue < s.min {
			newValue = s.max
		}
		if newValue > s.max {
			newValue = s.min
		}
		if oldValue != newValue {
			s.value[s.cursor] = newValue
			s.update = true
		}
		return int(newValue)
	}
	return value
}
