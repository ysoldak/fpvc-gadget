package main

import (
	"machine"
	"time"

	rotary_encoder "fpvc-gadget/src/rotary"
)

type Encoder struct {
	device *rotary_encoder.Device

	clickHandler  func()
	changeHandler func(value int) int

	lastClick time.Time
}

func NewEncoder() *Encoder {
	return &Encoder{
		device:    rotary_encoder.New(machine.D2, machine.D1),
		lastClick: time.Now(),
	}
}

func (e *Encoder) Configure() {
	e.device.Configure()
	machine.D0.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
}

func (e *Encoder) SetClickHandler(handler func()) {
	e.clickHandler = handler
	machine.D0.SetInterrupt(machine.PinRising, nil)
	if handler == nil {
		return
	}
	machine.D0.SetInterrupt(machine.PinRising, func(machine.Pin) {
		if time.Since(e.lastClick) > 100*time.Millisecond && e.clickHandler != nil {
			e.clickHandler()
		}
		e.lastClick = time.Now()
	})
}

func (e *Encoder) SetChangeHandler(handler func(value int) int, initValue int) {
	e.device.SetValue(initValue)
	e.changeHandler = handler
}

func (e *Encoder) Run() {
	oldValue := 0
	for {
		time.Sleep(100 * time.Microsecond)
		newValue := e.device.Value()
		if newValue == oldValue {
			continue
		}
		if e.changeHandler != nil {
			tmpValue := e.changeHandler(newValue)
			if tmpValue != newValue {
				newValue = tmpValue
				e.device.SetValue(newValue)
			}
		}
		oldValue = newValue
	}

}

func (e *Encoder) ClearHandlers() {
	e.SetClickHandler(func() {})
	e.SetChangeHandler(func(int) int { return 0 }, 0)
}
