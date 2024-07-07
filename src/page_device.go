package main

import (
	"fmt"
	"strings"
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

	pd.Draw()

	// Show info while fetching config
	display.Print(10, 24+20, "Fetching...")
	display.Show()

	// Do fetch
	display.Print(120, 60, "*")
	display.Show()
	err := settings.Fetch(pd.id)
	display.Erase(120, 60, "*")
	display.Show()
	if err != nil {
		display.Fill(10, 24+10, 128, 20, BLACK)
		display.Print(10, 24+20, err.Error())
		display.Show()
		pd.Page.redraw = false
		pd.Page.Enter()
		return
	}

	pd.cycler = func(iter int) {
		if pd.redraw { // after return from editing of an item, that may have changed number of items
			pd.items = pd.ItemsFromSettings()
		}
	}

	pd.ItemsFromSettings()
	pd.Draw()
	pd.Page.Enter()
}

func (pd *PageDevice) ItemsFromSettings() []Item {
	items := []Item{}
	items = append(items, NewItemSimple("- BATTLE ---------"))
	items = append(items, NewItemByte("Team", 72, 0x0A, 0x0E, 1).WithDrawer(&TeamNameDrawer{}))
	items = append(items, NewItemByte("Player", 73, 1, 9, 1))
	items = append(items, NewItemByte("Life", 70, 1, 255, 1))
	items = append(items, NewItemByte("Ammo", 71, 1, 255, 1))
	items = append(items, NewItemByte("Shoot Power", 75, 1, 10-settings.Get(76)-settings.Get(77), 1))
	items = append(items, NewItemByte("Shoot Rate", 76, 1, 10-settings.Get(75)-settings.Get(77), 1))
	items = append(items, NewItemByte("Armor", 77, 1, 10-settings.Get(75)-settings.Get(76), 1))
	items = append(items, NewItemSimple("- DISPLAY --------"))
	items = append(items, NewItemString("Name", 100, 10))
	items = append(items, NewItemByte("Canvas", 95, 0, 4, 1).WithDrawer(NewNamesDrawer("30x16", "50x18", "30x16C", "60x22", "53x20")).WithValuer(&BitsValuer{0b00011100}))
	items = append(items, NewItemByte("Icons", 95, 0, 2, 1).WithDrawer(NewNamesDrawer("Off", "Short", "Full")).WithValuer(&BitsValuer{0b00000011}))
	items = append(items, NewItemByte("Debug", 96, 0, 1, 1).WithDrawer(NewNamesDrawer("Off", "On")).WithValuer(&BitsValuer{0b00000011}))
	items = append(items, NewItemSimple("- EFFECTS --------"))
	items = append(items, NewItemByte("Alive", 80, 0, 1, 1).WithDrawer(NewNamesDrawer("Pass", "Custom")))
	if settings.Get(80) == 1 { // Custom
		items = append(items, NewItemByte(" Custom", 81, 0, 24, 1).WithDrawer(&LinearDrawer{50, 900}))
	}
	items = append(items, NewItemByte("Dead", 82, 0, 1, 1).WithDrawer(NewNamesDrawer("Pass", "Custom")))
	if settings.Get(82) == 1 { // Custom
		items = append(items, NewItemByte(" Custom", 83, 0, 24, 1).WithDrawer(&LinearDrawer{50, 900}))
	}
	items = append(items, NewItemByte("Hit", 84, 0, 2, 1).WithDrawer(NewNamesDrawer("Pass", "Custom", "Shake")))
	if settings.Get(84) != 0 { // !Pass
		items = append(items, NewItemByte(" Duration MS", 85, 1, 120, 1).WithDrawer(&LinearDrawer{200, 0}))
	}
	if settings.Get(84) == 1 { // Custom
		items = append(items, NewItemByte(" Custom", 86, 0, 24, 1).WithDrawer(&LinearDrawer{50, 900}))
	}
	if settings.Get(84) == 2 { // Shake
		items = append(items, NewItemByte(" Shake Every", 87, 1, 10, 1).WithDrawer(&LinearDrawer{200, 0}))
		items = append(items, NewItemByte(" Shake Level", 88, 1, 12, 1).WithDrawer(&LinearDrawer{50, 0}))
	}
	items = append(items, NewItemSimple("- HARDWARE -------"))
	items = append(items, NewItemByte("Team Leds", 91, 0, 255, 1))
	items = append(items, NewItemByte("Info Leds", 92, 0, 50, 2))
	items = append(items, NewItemByte("Voltage", 93, 1, 255, 1))

	if strings.Contains(pd.device.Hardware, "2.6") {
		if strings.HasPrefix(pd.device.Firmware, "2.7") {
			items = append(items, NewItemByte("HC12 Pins", 94, 0, 3, 1).WithDrawer(NewNamesDrawer("Off", "CSP", "MSP AU", "MSP FC")).WithValuer(&BitsValuer{0b00000011}))
			items = append(items, NewItemByte("I2C  Pins", 94, 0, 1, 1).WithDrawer(NewNamesDrawer("I2C", "CSP")).WithValuer(&BitsValuer{0b00110000}))
			items = append(items, NewItemByte("MSP  Pins", 94, 0, 3, 1).WithDrawer(NewNamesDrawer("Auto", "Off", "MSP AU", "MSP FC")).WithValuer(&BitsValuer{0b00001100}))
		} else {
			items = append(items, NewItemByte("MSP  Pins", 94, 0, 2, 1).WithDrawer(NewNamesDrawer("Auto", "Off", "MSP AU")).WithValuer(&BitsValuer{0b00001100}))
		}
	}
	return items
}
