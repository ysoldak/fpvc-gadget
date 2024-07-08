package main

import (
	"fmt"
	"machine"
	"time"

	csp "github.com/ysoldak/fpvc-serial-protocol"
)

var Version string

var battery *Battery
var display *Display

var encoder *Encoder
var network *csp.Adapter

func main() {

	battery = NewBattery()
	battery.Configure()

	serial := NewSerial(machine.UART0, machine.D7, machine.D6)
	serial.Configure()
	// csp.Logger = machine.Serial // debug serial communication
	network = csp.NewAdapter(serial.uart)

	display = &Display{}
	display.Configure()

	time.Sleep(time.Second)

	display.device.ClearDisplay()
	display.Print(16, 10, "FPV Combat Gadget")
	display.Print(16, 24, Version)
	display.Print(16, 50, fmt.Sprintf("Battery: %3.2fV", battery.Voltage()))
	display.device.Display()

	time.Sleep(3 * time.Second)

	encoder = NewEncoder()
	encoder.Configure()
	go encoder.Run()

	scanPage := NewPageScan()
	// scanPage.items = append(scanPage.items, &DeviceItem{Id: 0xA1, Name: "Alice", Firmware: "2.6.2", Hardware: "2.5"})
	// scanPage.items = append(scanPage.items, &DeviceItem{Id: 0xB2, Name: "Bob", Firmware: "2.6.2", Hardware: "2.6"})

	scanPage.Enter()

}
