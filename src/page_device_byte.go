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

func (ib *ItemByte) WithDrawer(drawer ItemDrawer) *ItemByte {
	ib.ItemDrawer = drawer
	return ib
}

func (ib *ItemByte) WithValuer(valuer ItemValuer) *ItemByte {
	ib.ItemValuer = valuer
	return ib
}

func (ib *ItemByte) Draw(row int) {
	ib.row = row
	display.Fill(20, pageBodyOffset+int16(row)*pageRowHeight-8, 100, 10, BLACK)
	ib.ItemDrawer.Draw(row, ib)
}

func (ib *ItemByte) Enter() {
	ib.editing = true
	encoder.SetClickHandler(ib.HandleClick)
	encoder.SetChangeHandler(ib.HandleChange, int(ib.getValue(ib.address)))
	display.Print(120, pageBodyOffset+int16(ib.row)*pageRowHeight, "<")
	display.Show()
	for ib.editing {
		if ib.changed {
			ib.changed = false
			ib.Draw(ib.row)
			display.Show()
		}
		time.Sleep(10 * time.Millisecond)
	}
	display.Print(120, 60, "*")
	display.Show()
	err := settings.Commit(ib.address, 1)
	display.Erase(120, 60, "*")
	display.Show()
	if err != nil {
		println(err.Error())
	}
}

func (ib *ItemByte) HandleClick() {
	ib.editing = false
}

func (ib *ItemByte) HandleChange(value int) int {
	eValue := ib.ItemValuer.getValue(ib.address)
	diff := (value - int(eValue)) * int(ib.inc)
	value = int(eValue) + diff
	if value < int(ib.min) {
		value = int(ib.max)
	}
	if value > int(ib.max) {
		value = int(ib.min)
	}
	ib.ItemValuer.setValue(ib.address, byte(value))
	ib.changed = eValue != byte(value)
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
