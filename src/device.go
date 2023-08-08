package main

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrTimeout     = errors.New("Request Timeout")
	ErrWrongId     = errors.New("Request Wrong ID")
	ErrWrongVerson = errors.New("Request Wrong Version")
	ErrWrongCrc    = errors.New("Request Wrong Crc")
	ErrSetFailed   = errors.New("Request Set Failed")
)

var buf []byte = make([]byte, 1000)

type Device struct {
	id   byte
	name string
	life byte
	ammo byte

	eeprom     []byte
	eepromPrev []byte

	cursor int
	active bool
}

func (d *Device) Open() {

	d.active = true

	encoder.SetChangeHandler(nil)

	d.cursor = 4
	encoder.SetClickHandler(d.HandleClick)

	display.device.ClearDisplay()

	display.Print(10, 10, fmt.Sprintf("%X", d.id))

	display.Print(10, 30, "Refreshing...")

	display.Print(10, 60, "> Back")

	display.device.Display()

	err := d.Get(true)
	if err != nil {
		println(err.Error())
		display.Clear(10, 30, "Refreshing...")
		display.Print(10, 30, err.Error())
		display.device.Display()
	} else {
		d.Show()
		encoder.SetChangeHandler(d.HandleChange)
		encoder.SetClickHandler(d.HandleClick)
	}

	for d.active {
		time.Sleep(100 * time.Millisecond)
	}
}

func (d *Device) Show() {
	display.device.ClearDisplay()

	display.Print(10, 10, fmt.Sprintf("%X", d.id))

	display.Print(10, 20, fmt.Sprintf("  ID:   %X", d.id))
	display.Print(10, 30, fmt.Sprintf("  Name: %s", d.name))
	display.Print(10, 40, fmt.Sprintf("  Life: %d", d.life))
	display.Print(10, 50, fmt.Sprintf("  Ammo: %d", d.ammo))

	display.Print(10, 60, "  Back")

	display.Print(10, 20+int16(d.cursor)*10, ">")

	display.device.Display()
}

func (d *Device) HandleClick() {
	if d.cursor == 4 {
		d.active = false
	}
	// TODO handle other positions
}

func (d *Device) HandleChange(value int) int {
	orig := d.cursor

	d.cursor = value
	if d.cursor < 0 {
		d.cursor = 4
	}
	if d.cursor > 4 {
		d.cursor = 0
	}

	if orig != d.cursor {
		d.Show()
	}
	return d.cursor
}

func (d *Device) Get(send bool) error {
	if send {
		serial.uart.WriteByte(0x72)
		serial.uart.WriteByte(d.id)
	}

	time.Sleep(1 * time.Second)

	i := 0
	for serial.uart.Buffered() < 113 {
		if i > 50 {
			return ErrTimeout
		}
		time.Sleep(100 * time.Millisecond)
		i++
	}

	_, err := serial.uart.Read(buf)
	if err != nil {
		return err
	}

	if buf[0] != d.id {
		return ErrWrongId
	}

	if buf[1] != 1 {
		return ErrWrongVerson
	}

	crc := byte(0)
	for i := 0; i < 112; i++ {
		crc += buf[i]
	}
	if buf[112] != crc {
		return ErrWrongCrc
	}

	println(len(d.eeprom))

	if len(d.eeprom) == 0 {
		d.eeprom = make([]byte, 110)
	}

	copy(d.eeprom, buf[2:112])

	println(len(d.eeprom))
	d.life = d.eeprom[62]
	println("1")
	d.ammo = d.eeprom[63]

	println("2")
	println(len(d.eeprom))

	return nil
}

func (d *Device) Set() error {

	// send new config
	serial.uart.WriteByte(0x74)
	serial.uart.WriteByte(d.id)
	serial.uart.WriteByte(1)
	serial.uart.Write(d.eeprom)
	crc := byte(0)
	for i := 0; i < 112; i++ {
		crc += d.eeprom[i]
	}
	serial.uart.WriteByte(crc)

	if len(d.eepromPrev) == 0 {
		d.eepromPrev = make([]byte, 110)
	}
	copy(d.eepromPrev, d.eeprom)

	// wait confirmation
	err := d.Get(false)
	if err != nil {
		return err
	}

	for i := 0; i < 110; i++ {
		if d.eepromPrev[i] != d.eeprom[i] {
			return ErrSetFailed
		}
	}

	return nil
}
