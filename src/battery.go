package main

import "machine"

type Battery struct {
	pinVoltage machine.Pin
	pinRead    machine.Pin
	adc        machine.ADC
}

func NewBattery() *Battery {
	return &Battery{
		pinRead:    machine.P0_14,
		pinVoltage: machine.P0_31,
	}
}

func (b *Battery) Configure() {
	// Enable charging at high current, 100mA
	machine.P0_13.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.P0_13.Low()

	// Shall keep this low while reading voltage
	b.pinRead.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Battery sensor pin
	b.adc = machine.ADC{Pin: b.pinVoltage}
	b.adc.Configure(machine.ADCConfig{})
}

func (b *Battery) Voltage() float64 {
	b.pinRead.Low()
	return float64(b.adc.Get()) / 7100
}
