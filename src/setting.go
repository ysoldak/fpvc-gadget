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

	min byte // min value of each byte
	max byte // max value of each byte

	cursor byte
	len    byte // total length of the value slice

	// visualisation
	kind     SettingKind
	show     byte
	title    string
	position byte
	update   bool
	active   bool

	positionOffset byte
}

func (s *Setting) Open() {

	s.active = true
	s.cursor = 0

	encoder.SetClickHandler(s.HandleClick)
	encoder.SetChangeHandler(s.HandleChange)
	encoder.device.SetValue(int(s.value[s.cursor]))

	for s.active {
		if s.update {
			s.update = false
			s.Show()
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func (s *Setting) Show() {
	display.Rect(30, int16(s.positionOffset+10*(s.position-1)+2), 100, 10, BLACK)
	switch s.show {
	case SettingShowDec:
		display.Print(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %d", s.title, s.value)) // TODO support global scroll
	case SettingShowHex:
		display.Print(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %X", s.title, s.value)) // TODO support global scroll
	case SettingShowChar:
		display.Print(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %s", s.title, string(s.value))) // TODO support global scroll
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
