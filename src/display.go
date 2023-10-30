package main

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

var BLACK = color.RGBA{0, 0, 0, 255}
var WHITE = color.RGBA{255, 255, 255, 255}

type Display struct {
	device ssd1306.Device
}

func (d *Display) Configure() {
	machine.I2C1.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       machine.D4,
		SCL:       machine.D5,
	})
	d.device = ssd1306.NewI2C(machine.I2C1)
	d.device.Configure(ssd1306.Config{
		Address: 0x3C,
		Width:   128,
		Height:  64,
	})
	d.device.ClearDisplay()
}

func (d *Display) Print(x, y int16, message string) {
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, x, y, message, WHITE, tinyfont.NO_ROTATION)
}

func (d *Display) Clear(x, y int16, message string) {
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, x, y, message, BLACK, tinyfont.NO_ROTATION)
}

func (d *Display) Fill(x, y, w, h int16, color color.RGBA) {
	tinydraw.FilledRectangle(&d.device, x, y, w, h, color)
}

func (d *Display) Rect(x, y, w, h int16, color color.RGBA) {
	tinydraw.Rectangle(&d.device, x, y, w, h, color)
}

func (d *Display) Line(x0, y0, x1, y1 int16, color color.RGBA) {
	tinydraw.Line(&d.device, x0, y0, x1, y1, color)
}
