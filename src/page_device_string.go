package main

import (
	"fmt"
	"time"
)

type ItemString struct {
	title   string
	address byte
	length  byte

	row     int
	editing bool
	changed bool

	pos byte
}

func NewItemString(title string, address, length byte) *ItemString {
	return &ItemString{
		title:   title,
		address: address,
		length:  length,
	}
}

func (is *ItemString) Draw(row int) {
	is.row = row
	str := ""
	for i := byte(0); i < is.length; i++ {
		str += string(settings.Get(is.address + i))
	}
	if is.editing {
		display.Fill(2+6, pageBodyOffset+int16(row)*pageRowHeight-8, 128-8-8, 12, BLACK)
		charX := 2 + int16(1+len(is.title)+2)*6 + int16(is.pos)*6
		display.Line(charX, pageBodyOffset+int16(row)*pageRowHeight+2, charX+6, pageBodyOffset+int16(row)*pageRowHeight+2, WHITE)
	}
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %s", is.title, str))
}

func (is *ItemString) Enter() {
	is.pos = 0
	is.editing = true
	encoder.SetClickHandler(is.HandleClick)
	encoder.SetChangeHandler(is.HandleChange, int(settings.Get(is.address)))
	display.Print(120, pageBodyOffset+int16(is.row)*pageRowHeight, "<")
	is.Draw(is.row)
	display.Show()
	tmpPos := is.pos
	for is.editing {
		if is.changed || tmpPos != is.pos {
			tmpPos = is.pos
			is.changed = false
			is.Draw(is.row)
			display.Show()
		}
		time.Sleep(10 * time.Millisecond)
	}
	err := settings.Commit()
	if err != nil {
		println(err.Error())
	}
}

func (is *ItemString) HandleClick() {
	is.pos++
	if is.pos == is.length {
		is.editing = false
		return
	}
	encoder.SetValue(int(settings.Get(is.address + is.pos))) // set encoder to current active char value
}

func (is *ItemString) HandleChange(value int) int {
	eValue := settings.Get(is.address + is.pos)
	switch {
	case value < 0x20:
		value = 0x5A
	case value > 0x20 && value < 0x30:
		if eValue < byte(value) {
			value = 0x30
		} else {
			value = 0x20
		}
	case value > 0x5A:
		value = 0x20
	}
	settings.Set(is.address+is.pos, byte(value))
	is.changed = eValue != byte(value)
	return value
}
