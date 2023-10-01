package main

import (
	"fmt"
	"machine"
	"time"
)

var Version string

var battery *Battery
var display *Display
var serial *Serial
var encoder *Encoder

var scan *Scan

func main() {

	battery = NewBattery()
	battery.Configure()

	serial = NewSerial(machine.UART0, machine.D7, machine.D6)
	serial.Configure()

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

	scan = &Scan{}
	// scan.devices = append(scan.devices, &Device{id: 0xB2, name: "Bob"})
	// scan.devices = append(scan.devices, &Device{id: 0xD4, name: "Dude"})

	scan.Open()

}
