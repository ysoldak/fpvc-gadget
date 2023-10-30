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

	scroll  int
	cursor  int
	active  bool
	changed bool
	failure bool
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

	d.cursor = 0
	encoder.SetClickHandler(d.HandleClick)

	display.device.ClearDisplay()

	display.Print(10, 10, fmt.Sprintf("%X", d.id))

	display.Print(10, 30, "Refreshing...")

	display.Print(10, 60, "> Back")

	display.device.Display()

	err := d.Get(true)
	if err != nil {
		d.failure = true
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
			if d.cursor == len(d.settings)+1 { // save & back
				if d.changed {
					display.Clear(10, 10, fmt.Sprintf("%X !!!", d.id))
					display.Print(10, 10, fmt.Sprintf("%X", d.id))
					display.device.Display()
					err := d.Set()
					if err != nil {
						display.Print(10, 10, fmt.Sprintf("%X !!!", d.id))
						display.device.Display()
						d.active = true
						continue
					}
				}
				return
			}
			if d.failure || d.cursor == len(d.settings)+2 { // failure to fetch config OR cancel & back
				return
			}

			display.Clear(10, 20+int16(d.cursor-d.scroll)*10, ">")
			display.Print(10, 20+int16(d.cursor-d.scroll)*10, "*")
			display.device.Display()

			setting := d.settings[d.cursor]
			setting.Open(d.cursor - d.scroll)
			if d.cursor == 0 { // special handling for ID setting
				d.eeprom[setting.address] = setting.value[0] >> 4
				d.eeprom[setting.address+1] = setting.value[0] & 0x0F
			} else {
				for i := byte(0); i < setting.len; i++ {
					d.eeprom[setting.address+i] = setting.value[i]
				}
			}

			display.Clear(10, 20+int16(d.cursor-d.scroll)*10, "*")
			display.Print(10, 20+int16(d.cursor-d.scroll)*10, ">")
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

	for i, s := range d.settings {
		if i < d.scroll || i > d.scroll+4 {
			continue
		}
		s.Show(i - d.scroll)
	}

	display.Print(10, int16(20+(len(d.settings)-d.scroll)*10), "  ----------------")
	display.Print(10, int16(20+(len(d.settings)+1-d.scroll)*10), "  Save & Back")
	display.Print(10, int16(20+(len(d.settings)+2-d.scroll)*10), "  Cancel & Back")

	display.Print(10, 20+int16(d.cursor-d.scroll)*10, ">")

	display.device.Display()
}

func (d *Device) HandleClick() {
	d.active = false
}

func (d *Device) HandleChange(value int) int {
	orig := d.cursor

	d.cursor = value
	if d.cursor < 0 {
		d.cursor = 0
	}
	if d.cursor == len(d.settings) { // divider
		if orig < d.cursor {
			d.cursor++
		} else {
			d.cursor--
		}
	}
	if d.cursor > len(d.settings)+2 {
		d.cursor = len(d.settings) + 2
	}
	scrollChanged := false
	if d.cursor-d.scroll < 0 {
		d.scroll--
		scrollChanged = true
	}
	if d.cursor-d.scroll > 4 {
		d.scroll += d.cursor - d.scroll - 4
		scrollChanged = true
	}

	if scrollChanged {
		d.Show()
	} else {
		if orig != d.cursor {
			display.Clear(10, 20+int16(orig-d.scroll)*10, ">")
			display.Print(10, 20+int16(d.cursor-d.scroll)*10, ">")
			display.device.Display()
		}
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
		address: 64,
		value:   []byte{(d.eeprom[64] << 4) + d.eeprom[65]},
		min:     0xA1,
		max:     0xF9,
		invalid: []byte{
			0xA0, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF,
			0xB0, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF,
			0xC0, 0xCA, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF,
			0xD0, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF,
			0xE0, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF,
			0xF0, 0xFA, 0xFB, 0xFC, 0xFD, 0xFE, 0xFF,
		},
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowHex,
		title:          "ID",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        100,
		value:          make([]byte, 10),
		min:            ' ',
		max:            'z',
		len:            10,
		kind:           SettingKindByte,
		show:           SettingShowChar,
		title:          "Name",
		positionOffset: 20,
	})
	copy(d.settings[1].value, d.eeprom[100:110])

	d.settings = append(d.settings, Setting{
		address:        62,
		value:          []byte{d.eeprom[62]},
		min:            0,
		max:            255,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Life",
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
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        69,
		value:          []byte{d.eeprom[69]},
		min:            1,
		max:            8,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Shoot Power",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        70,
		value:          []byte{d.eeprom[70]},
		min:            1,
		max:            8,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Shoot Rate",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        71,
		value:          []byte{d.eeprom[71]},
		min:            1,
		max:            8,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Defense",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        73,
		value:          []byte{d.eeprom[73]},
		min:            0,
		max:            255,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Shake Speed",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        74,
		value:          []byte{d.eeprom[74]},
		min:            0,
		max:            255,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Shake Level",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        80,
		value:          []byte{d.eeprom[80]},
		min:            0,
		max:            255,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Effect Time",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        90,
		value:          []byte{d.eeprom[90]},
		min:            0,
		max:            2,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Output Mode",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        91,
		value:          []byte{d.eeprom[91]},
		min:            0,
		max:            255,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Team Led",
		positionOffset: 20,
	})

	d.settings = append(d.settings, Setting{
		address:        92,
		value:          []byte{d.eeprom[92]},
		min:            0,
		max:            255,
		len:            1,
		kind:           SettingKindByte,
		show:           SettingShowDec,
		title:          "Info Led",
		positionOffset: 20,
	})

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
