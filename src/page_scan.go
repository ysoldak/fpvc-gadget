package main

import (
	"fmt"
	"strings"
	"time"
)

type PageScan struct {
	Page
}

func NewPageScan() *PageScan {
	ps := &PageScan{
		Page: Page{
			Title: "Scanning...",
			items: []Item{},
		},
	}
	ps.cycler = ps.Cycle
	return ps
}

func (ps *PageScan) Cycle(iter int) {
	if iter%500 == 0 { // every 5 sec
		err := serial.uart.WriteByte(0x70)
		if err != nil {
			println(err.Error())
		}
		display.Print(120, 60, "*")
		display.Show()
	}
	if (iter-100)%500 == 0 { // every 5+1 sec
		display.Erase(120, 60, "*")
		display.Show()
	}
	ps.Receive()
}

func (ps *PageScan) Receive() {
	if serial.uart.Buffered() < 32 {
		return
	}
	key := byte(0)
	name := make([]byte, 10)
	desc := make([]byte, 20)
	for i := 0; i < 32; i++ {
		b, err := serial.uart.ReadByte()
		if err != nil {
			return
		}
		switch {
		case i == 0:
			// fmt.Printf("Receiving %X\n", b)
			if !(0xA0 < b && b < 0xFF) {
				return
			}
			key = b
		case 1 < i && i < 12:
			if b == 0x2F {
				b = 0x20
			}
			name[i-2] = b
		case 11 < i && i < 32:
			if b == 0x2F {
				b = 0x20
			}
			desc[i-12] = b
		}
	}
	new := true
	device := &DeviceItem{}
	for _, item := range ps.items {
		di := item.(*DeviceItem)
		if di.Id == key {
			new = false
			device = di
			break
		}
	}
	device.Id = key
	device.Name = string(name)
	device.Firmware = strings.TrimSpace(strings.Split(string(desc), " ")[0])
	device.Hardware = strings.TrimSpace(strings.Split(string(desc), " ")[1])
	device.lastSeen = time.Now()
	if new {
		ps.items = append(ps.items, device)
		ps.redraw = true
	}
	fmt.Printf("Scan: + %X | %s | %s | %s\r\n", device.Id, device.Name, device.Firmware, device.Hardware)

	// remove old devices
	for i, item := range ps.items {
		di := item.(*DeviceItem)
		if time.Since(di.lastSeen) > 10*time.Second {
			fmt.Printf("Scan: - %X | %s | %s | %s\r\n", di.Id, di.Name, di.Firmware, di.Hardware)
			ps.items = append(ps.items[:i], ps.items[i+1:]...)
			ps.redraw = true
		}
	}
	if ps.cursor >= len(ps.items) {
		ps.cursor = len(ps.items) - 1
	}

}

// ----------------------------

type DeviceItem struct {
	Id       byte
	Name     string
	Firmware string
	Hardware string
	lastSeen time.Time
}

func (di *DeviceItem) Enter() {
	pd := NewPageDevice(di)
	pd.Enter()
}

func (di *DeviceItem) Draw(row int) {
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, fmt.Sprintf(" %X | %s ", di.Id, di.Name))
}
