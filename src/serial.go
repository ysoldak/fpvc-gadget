package main

import "machine"

type Serial struct {
	uart *machine.UART
	rx   machine.Pin
	tx   machine.Pin
}

func NewSerial(uart *machine.UART, rx, tx machine.Pin) *Serial {
	return &Serial{
		uart: uart,
		rx:   rx,
		tx:   tx,
	}
}

func (s *Serial) Configure() {
	s.uart.Configure(machine.UARTConfig{TX: s.tx, RX: s.rx, BaudRate: 9600})
}
