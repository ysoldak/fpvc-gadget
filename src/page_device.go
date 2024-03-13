package main

import (
	"fmt"
	"strings"
	"time"
)

type PageDevice struct {
	Page
	id     byte
	device *DeviceItem
}

func NewPageDevice(item *DeviceItem) *PageDevice {
	pd := &PageDevice{
		Page: Page{
			Title:   fmt.Sprintf("%-10s %10s", strings.TrimSpace(item.Name), strings.TrimSpace(item.Firmware)),
			items:   []Item{},
			subpage: true,
			redraw:  true,
		},
		id:     item.Id,
		device: item,
	}
	return pd
}

func (pd *PageDevice) Enter() {

	encoder.ClearHandlers()

	// Show info while fetching config
	display.Print(10, 24+20, "Fetching...")
	display.Show()

	// Do fetch
	err := settings.Fetch(pd.id)
	if err != nil {
		display.Fill(10, 24+10, 128, 20, BLACK)
		display.Print(10, 24+20, err.Error())
		display.Show()
		pd.Page.Enter()
		return
	}

	pd.ItemsFromEeprom()
	pd.Draw()
	pd.Page.Enter()
}

func (pd *PageDevice) ItemsFromEeprom() {
	pd.items = []Item{}
	pd.items = append(pd.items, NewItemSimple("- BATTLE ---------"))
	pd.items = append(pd.items, NewSettingItem("Team", 72, 0x0A, 0x0E, 1).WithDrawer(&TeamNameDrawer{}))
	pd.items = append(pd.items, NewSettingItem("Life", 70, 1, 255, 1))
	pd.items = append(pd.items, NewSettingItem("Ammo", 71, 1, 255, 1))
	pd.items = append(pd.items, NewItemSimple("- DISPLAY --------"))
	pd.items = append(pd.items, NewSettingItem("Canvas", 95, 0, 4, 1).WithDrawer(&CanvasSizeDrawer{}).WithHandler(&BitsHandler{0b00011100}))
	pd.items = append(pd.items, NewItemSimple("- HARDWARE -------"))
	pd.items = append(pd.items, NewSettingItem("Team Leds", 91, 0, 255, 1))
	pd.items = append(pd.items, NewSettingItem("Info Leds", 92, 0, 50, 2))
	pd.items = append(pd.items, NewSettingItem("Voltage", 93, 1, 255, 1))
	pd.items = append(pd.items, NewSettingItem("HC12 pins", 94, 0, 1, 1).WithDrawer(&HC12PortDrawer{}).WithHandler(&BitsHandler{0b00000011}))
	if strings.Contains(pd.device.Hardware, "2.6") {
		pd.items = append(pd.items, NewSettingItem("MSP Pins", 94, 0, 2, 1).WithDrawer(&MSPPortDrawer{}).WithHandler(&BitsHandler{0b00001100}))
	}
}

// ----------------------------------------

type SettingItemDrawer interface {
	Draw(row int, si *SettingItem)
}

type SettingValueHandler interface {
	getValue(address byte) byte
	setValue(address byte, value byte)
}

type SettingItem struct {
	SettingItemDrawer
	SettingValueHandler

	Title string

	address byte

	min byte
	max byte
	inc byte

	row     int
	editing bool
	changed bool
}

func NewSettingItem(title string, address, min, max, inc byte) *SettingItem {
	return &SettingItem{
		Title:               title,
		address:             address,
		min:                 min,
		max:                 max,
		inc:                 inc,
		SettingItemDrawer:   &DefaultDrawer{},
		SettingValueHandler: &DefaultHandler{},
	}
}

func (si *SettingItem) WithDrawer(drawer SettingItemDrawer) *SettingItem {
	si.SettingItemDrawer = drawer
	return si
}

func (si *SettingItem) WithHandler(handler SettingValueHandler) *SettingItem {
	si.SettingValueHandler = handler
	return si
}

func (si *SettingItem) Draw(row int) {
	si.row = row
	display.Fill(20, pageBodyOffset+int16(row)*pageRowHeight-8, 100, 10, BLACK)
	si.SettingItemDrawer.Draw(row, si)
}

func (si *SettingItem) Enter() {
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

func (si *SettingItem) HandleClick() {
	si.editing = false
}

func (si *SettingItem) HandleChange(value int) int {
	eValue := si.SettingValueHandler.getValue(si.address)
	if value < int(si.min) {
		value = int(si.max)
	}
	if value > int(si.max) {
		value = int(si.min)
	}
	si.SettingValueHandler.setValue(si.address, byte(value))
	si.changed = eValue != byte(value)
	return value
}

// ----------------------------------------

type DefaultHandler struct{}

func (dh *DefaultHandler) getValue(address byte) byte {
	return settings.Get(address)
}

func (dh *DefaultHandler) setValue(address byte, value byte) {
	settings.Set(address, value)
}

// ----------------------------------------

type BitsHandler struct {
	mask byte
}

func (bh *BitsHandler) getValue(address byte) byte {
	return settings.GetBits(address, bh.mask)
}

func (bh *BitsHandler) setValue(address byte, value byte) {
	settings.SetBits(address, bh.mask, value)
}

// ----------------------------------------

type DefaultDrawer struct{}

func (dd *DefaultDrawer) Draw(row int, si *SettingItem) {
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %d", si.Title, si.getValue(si.address)))
}

// ----------------------------------------

type TeamNameDrawer struct{}

func (tnd *TeamNameDrawer) Draw(row int, si *SettingItem) {
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %s", si.Title, string('A'+si.getValue(si.address)-0x0A)))
}

// ----------------------------------------

type CanvasSizeDrawer struct{}

func (csd *CanvasSizeDrawer) Draw(row int, si *SettingItem) {

	value := si.SettingValueHandler.getValue(si.address)
	valueStr := "Unknown"
	switch value {
	case 0:
		valueStr = "30x16"
	case 1:
		valueStr = "50x18"
	case 2:
		valueStr = "30x16C"
	case 3:
		valueStr = "60x22"
	case 4:
		valueStr = "53x20"
	}

	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %s", si.Title, valueStr))
}

type HC12PortDrawer struct{}

func (hd *HC12PortDrawer) Draw(row int, si *SettingItem) {
	value := si.SettingValueHandler.getValue(si.address)
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %s", si.Title, []string{"OFF", "NETWORK"}[value]))
}

type MSPPortDrawer struct{}

func (hd *MSPPortDrawer) Draw(row int, si *SettingItem) {
	value := si.SettingValueHandler.getValue(si.address)
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %s: %s", si.Title, []string{"AUTO", "OFF", "MSP OUT"}[value]))
}
