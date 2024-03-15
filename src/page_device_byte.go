package main

import (
	"fmt"
	"time"
)

type ItemDrawer interface {
	Draw(row int, si *ItemByte)
}

type ItemValuer interface {
	getValue(address byte) byte
	setValue(address byte, value byte)
}

type ItemByte struct {
	ItemDrawer
	ItemValuer

	Title string

	address byte

	min byte
	max byte
	inc byte

	row     int
	editing bool
	changed bool
}

func NewItemByte(title string, address, min, max, inc byte) *ItemByte {
	return &ItemByte{
		Title:      title,
		address:    address,
		min:        min,
		max:        max,
		inc:        inc,
		ItemDrawer: &DefaultDrawer{},
		ItemValuer: &DefaultValuer{},
	}
}

func (si *ItemByte) WithDrawer(drawer ItemDrawer) *ItemByte {
	si.ItemDrawer = drawer
	return si
}

func (si *ItemByte) WithValuer(valuer ItemValuer) *ItemByte {
	si.ItemValuer = valuer
	return si
}

func (si *ItemByte) Draw(row int) {
	si.row = row
	display.Fill(20, pageBodyOffset+int16(row)*pageRowHeight-8, 100, 10, BLACK)
	si.ItemDrawer.Draw(row, si)
}

func (si *ItemByte) Enter() {
	si.editing = true
	encoder.SetClickHandler(si.HandleClick)
	encoder.SetChangeHandler(si.HandleChange, int(si.getValue(si.address)))
	display.Print(120, pageBodyOffset+int16(si.row)*pageRowHeight, "<")
	display.Show()
	for si.editing {
		if si.changed {
			si.changed = false
			si.Draw(si.row)
			display.Show()
		}
		time.Sleep(10 * time.Millisecond)
	}
	err := settings.Commit()
	if err != nil {
		println(err.Error())
	}
}

func (si *ItemByte) HandleClick() {
	si.editing = false
}

func (si *ItemByte) HandleChange(value int) int {
	eValue := si.ItemValuer.getValue(si.address)
	if value < int(si.min) {
		value = int(si.max)
	}
	if value > int(si.max) {
		value = int(si.min)
	}
	si.ItemValuer.setValue(si.address, byte(value))
	si.changed = eValue != byte(value)
	return value
}

// ----------------------------------------

type DefaultValuer struct{}

func (dh *DefaultValuer) getValue(address byte) byte {
	return settings.Get(address)
}

func (dh *DefaultValuer) setValue(address byte, value byte) {
	settings.Set(address, value)
}

// ----------------------------------------

type BitsValuer struct {
	mask byte
}

func (bh *BitsValuer) getValue(address byte) byte {
	return settings.GetBits(address, bh.mask)
}

func (bh *BitsValuer) setValue(address byte, value byte) {
	settings.SetBits(address, bh.mask, value)
}

// ----------------------------------------

type DefaultDrawer struct{}

func (dd *DefaultDrawer) Draw(row int, si *ItemByte) {
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %d", si.Title, si.getValue(si.address)))
}

// ----------------------------------------

type TeamNameDrawer struct{}

func (tnd *TeamNameDrawer) Draw(row int, si *ItemByte) {
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %s", si.Title, string('A'+si.getValue(si.address)-0x0A)))
}

// ----------------------------------------

type NamesDrawer struct {
	names []string
}

func NewNamesDrawer(names ...string) *NamesDrawer {
	return &NamesDrawer{
		names: names,
	}
}

func (cd *NamesDrawer) Draw(row int, si *ItemByte) {
	value := si.getValue(si.address)
	str := "???"
	if int(value) < len(cd.names) {
		str = cd.names[value]
	}
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %s", si.Title, str))
}

// ----------------------------------------

type LinearDrawer struct {
	multiplier int
	offset     int
}

func NewLinearDrawer(a, b int) *LinearDrawer {
	return &LinearDrawer{
		multiplier: a,
		offset:     b,
	}
}

func (ld *LinearDrawer) Draw(row int, si *ItemByte) {
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %d", si.Title, ld.multiplier*int(si.getValue(si.address))+ld.offset))
}
