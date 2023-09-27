package main

import (
	"machine"
)

var Version string

var display *Display
var serial *Serial
var encoder *Encoder

var scan *Scan

func main() {

	serial = NewSerial(machine.UART0, machine.D7, machine.D6)
	serial.Configure()

	display = &Display{}
	display.Configure()

	encoder = NewEncoder()
	encoder.Configure()
	go encoder.Run()

	scan = &Scan{}
	// scan.devices = append(scan.devices, &Device{id: 0xB2, name: "Bob"})
	// scan.devices = append(scan.devices, &Device{id: 0xD4, name: "Dude"})

	scan.Open()

}
