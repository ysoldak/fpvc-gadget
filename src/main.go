package main

import (
	"machine"
	"time"
)

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
	scan.devices = append(scan.devices, &Device{id: 0xB2, name: "Bob"})
	scan.devices = append(scan.devices, &Device{id: 0xD4, name: "Dude"})

	for {
		device := scan.Open()
		device.Open()
		time.Sleep(time.Second)
	}

}

// func receive() {
// 	if serial.uart.Buffered() < 22 {
// 		return
// 	}
// 	key := byte(0)
// 	name := make([]byte, 10)
// 	for i := 0; i < 22; i++ {
// 		b, err := serial.uart.ReadByte()
// 		if err != nil {
// 			return
// 		}
// 		if i == 0 && 0xA0 < b && b < 0xFF {
// 			key = b
// 		}
// 		if 1 < i && i < 12 {
// 			if b == 0x2F {
// 				b = 0x20
// 			}
// 			name[i-2] = b
// 		}
// 	}
// 	if key != 0 {
// 		devices[key] = string(name)
// 	}
// }

// func scan2() {
// 	i := int(0)
// 	for {

// 		println(i)

// 		display.device.ClearDisplay()
// 		if i%2 == 0 {
// 			display.Print(120, 60, "*")
// 		}

// 		if i%5 == 0 {
// 			serial.uart.WriteByte(0x70)
// 		}

// 		receive()

// 		j := int16(0)
// 		for id, name := range devices {
// 			fmt.Printf("%X | '%s'\r\n", id, string(name))
// 			display.Print(10, 10+j*10, fmt.Sprintf("  %X | %s ", id, name))
// 			j++
// 		}

// 		// display.Print(10, 10+cursor*10, ">")

// 		display.Print(10, 60, "FPV Combat Gadget")

// 		display.device.Display()
// 		time.Sleep(1 * time.Second)

// 		i++
// 	}

// }
