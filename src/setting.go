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

	oldValue       []byte // TODO remove this field and clear with rect
	positionOffset byte
}

func (s *Setting) Open() {

	s.active = true
	s.cursor = 0

	s.oldValue = make([]byte, s.len)
	copy(s.oldValue, s.value)

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
	// TODO draw empty rect to cleanup old value
	//display.Rect(10, int16(s.positionOffset+10*s.position), 30, int16(s.positionOffset+10*s.position)+10, BLACK)
	switch s.show {
	case SettingShowDec:
		display.Clear(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %d", s.title, s.oldValue)) // TODO clear with rect
		display.Print(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %d", s.title, s.value))    // TODO support global scroll
	case SettingShowHex:
		display.Clear(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %X", s.title, s.oldValue)) // TODO clear with rect
		display.Print(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %X", s.title, s.value))    // TODO support global scroll
	case SettingShowChar:
		display.Clear(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %s", s.title, string(s.oldValue))) // TODO clear with rect
		display.Print(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %s", s.title, string(s.value)))    // TODO support global scroll
	}
	display.device.Display()
}

func (s *Setting) HandleClick() {
	s.cursor++
	if s.cursor >= s.len {
		s.active = false
	}
}

func (s *Setting) HandleChange(value int) int {
	if s.kind == SettingKindByte {
		orig := s.value[s.cursor]
		s.value[s.cursor] = byte(value)
		if s.value[s.cursor] < s.min {
			s.value[s.cursor] = s.max
		}
		if s.value[s.cursor] > s.max {
			s.value[s.cursor] = s.min
		}
		if orig != s.value[s.cursor] {
			s.oldValue[s.cursor] = orig
			s.update = true
		}
		return int(s.value[s.cursor])
	}
	return value
}
