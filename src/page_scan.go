package main

import (
	"fmt"
	"strings"
	"time"

	csp "github.com/ysoldak/fpvc-serial-protocol"
)

type PageScan struct {
	Page
}

func NewPageScan() *PageScan {
	ps := &PageScan{
		Page: *NewPage("Scanning..."),
	}
	ps.cycler = ps.Cycle
	return ps
}

func (ps *PageScan) Cycle(iter int) {
	if iter%500 == 0 { // every 5 sec
		display.Print(120, 60, "*")
		display.Show()
	}
	if (iter-100)%500 == 0 { // every 5+1 sec
		display.Erase(120, 60, "*")
		display.Show()
	}
	ps.Receive()
	ps.Cleanup()
}

// remove old devices
func (ps *PageScan) Cleanup() {
	for i, item := range ps.items {
		di := item.(*DeviceItem)
		if time.Since(di.lastSeen) > 12*time.Second {
			fmt.Printf("Scan: - %X | %s | %s | %s\r\n", di.Id, di.Name, di.Firmware, di.Hardware)
			ps.items = append(ps.items[:i], ps.items[i+1:]...)
			ps.redraw = true
		}
	}
	if ps.cursor >= len(ps.items) {
		ps.cursor = len(ps.items) - 1
	}
	if ps.cursor < 0 {
		ps.cursor = 0
	}
}

func (ps *PageScan) Receive() {

	for {
		message, err := network.Receive()
		if err != nil && err == csp.ErrNoData {
			return
		}
		if err != nil {
			continue
		}
		if message.Command != csp.CommandBeacon {
			continue
		}
		beacon := csp.NewBeaconFromMessage(message)
		new := true
		device := &DeviceItem{}
		for _, item := range ps.items {
			di := item.(*DeviceItem)
			if di.Id == beacon.ID {
				new = false
				device = di
				break
			}
		}
		device.Id = beacon.ID
		device.Name = beacon.Name
		desc := beacon.Description
		device.Firmware = strings.TrimSpace(strings.Split(string(desc), " ")[0])
		device.Hardware = strings.TrimSpace(strings.Split(string(desc), " ")[1])
		device.lastSeen = time.Now()
		if new {
			ps.items = append(ps.items, device)
			ps.redraw = true
		}
		fmt.Printf("Scan: + %X | %s | %s | %s\r\n", device.Id, device.Name, device.Firmware, device.Hardware)
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
