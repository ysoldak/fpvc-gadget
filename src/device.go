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

	settings []Setting

	eeprom     []byte
	eepromPrev []byte

	cursor  int
	active  bool
	changed bool
}

func NewDevice() *Device {
	return &Device{
		settings: []Setting{},
	}
}

func (d *Device) Open() {

	d.active = true
	d.changed = false

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
		encoder.device.SetValue(d.cursor)
	}

	for {
		if !d.active {
			if d.cursor == 4 {
				if d.changed {
					display.Clear(10, 10, fmt.Sprintf("%X !!!", d.id))
					display.Print(10, 10, fmt.Sprintf("%X", d.id))
					err := d.Set()
					if err != nil {
						display.Print(10, 10, fmt.Sprintf("%X !!!", d.id))
						d.active = true
						continue
					}
				}
				return
			}

			display.Clear(10, 20+int16(d.cursor)*10, ">")
			display.Print(10, 20+int16(d.cursor)*10, "*")
			display.device.Display()

			setting := d.settings[d.cursor]
			setting.Open()
			for i := byte(0); i < setting.len; i++ {
				d.eeprom[setting.address+i] = setting.value[i]
			}

			display.Clear(10, 20+int16(d.cursor)*10, "*")
			display.Print(10, 20+int16(d.cursor)*10, ">")
			display.device.Display()

			encoder.SetClickHandler(d.HandleClick)
			encoder.SetChangeHandler(d.HandleChange)
			encoder.device.SetValue(d.cursor)
			d.active = true
			d.changed = true
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (d *Device) Show() {
	display.device.ClearDisplay()

	display.Print(10, 10, fmt.Sprintf("%X", d.id))

	for _, s := range d.settings {
		s.Show()
	}

	display.Print(10, 60, "  Save & Back")

	display.Print(10, 20+int16(d.cursor)*10, ">")

	display.device.Display()
}

func (d *Device) HandleClick() {
	d.active = false
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
		display.Clear(10, 20+int16(orig)*10, ">")
		display.Print(10, 20+int16(d.cursor)*10, ">")
		display.device.Display()
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

	d.settings = []Setting{}
	d.settings = append(d.settings, Setting{
		address:        100,
		value:          make([]byte, 10),
		min:            ' ',
		max:            'z',
		len:            10,
		kind:           SettingKindByte,
		show:           SettingShowChar,
		title:          "Name",
		position:       0,
		positionOffset: 20,
	})
	copy(d.settings[0].value, d.eeprom[100:110])

	d.settings = append(d.settings, Setting{
		address:        62,
		value:          []byte{d.eeprom[62]},
		min:            0,
		max:            255,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Life",
		position:       1,
		positionOffset: 20,
	})
	d.settings = append(d.settings, Setting{
		address:        63,
		value:          []byte{d.eeprom[63]},
		min:            0,
		max:            255,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Ammo",
		position:       2,
		positionOffset: 20,
	})
	d.settings = append(d.settings, Setting{
		address:        64,
		value:          []byte{d.eeprom[64]},
		min:            0xA,
		max:            0xF,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowHex,
		title:          "Team",
		position:       3,
		positionOffset: 20,
	})
	// d.settings = append(d.settings, Setting{
	// 	address:        65,
	// 	value:          []byte{d.eeprom[65]},
	// 	min:            0x1,
	// 	max:            0xF,
	// 	len:            1,
	// 	kind:           SettingKindByte,
	// 	show:           SettingShowHex,
	// 	title:          "Player",
	// 	position:       4,
	// 	positionOffset: 20,
	// })

	return nil
}

func (d *Device) Set() error {

	// send new config
	serial.uart.WriteByte(0x74)
	serial.uart.WriteByte(d.id)
	serial.uart.WriteByte(1)
	serial.uart.Write(d.eeprom)
	crc := d.id + 1
	for i := 0; i < 110; i++ {
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
