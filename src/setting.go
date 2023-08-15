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

type Setting struct {
	address byte
	value   byte

	// visualisation
	kind     SettingKind
	title    string
	position byte
	update   bool
	active   bool

	oldValue       byte // TODO remove this field and clear with rect
	positionOffset byte
}

func (s *Setting) Open() {

	s.active = true

	encoder.SetClickHandler(s.HandleClick)
	encoder.SetChangeHandler(s.HandleChange)
	encoder.device.SetValue(int(s.value))

	for s.active {
		if s.update {
			s.update = false
			display.Clear(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %d", s.title, s.oldValue)) // TODO support global scroll; clear with rect
			//display.Rect(10, int16(s.positionOffset+10*s.position), 30, int16(s.positionOffset+10*s.position)+10, BLACK)
			// TODO draw empty rect to cleanup old value
			s.Show()
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func (s *Setting) Show() {
	display.Print(10, int16(s.positionOffset+10*s.position), fmt.Sprintf("  %s: %d", s.title, s.value)) // TODO support global scroll
	display.device.Display()
}

func (s *Setting) HandleClick() {
	s.active = false
}

func (s *Setting) HandleChange(value int) int {
	if s.kind == SettingKindByte {
		orig := s.value
		s.value = byte(value)
		if s.value < 0 {
			s.value = 255
		}
		if s.value > 255 {
			s.value = 0
		}
		if orig != s.value {
			s.oldValue = orig
			s.update = true
		}
		return int(s.value)
	}
	return value
}
