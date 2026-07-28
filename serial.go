//go:build serial.rtt

package main

import (
	"machine"
	"machine/usb/cdc"
)

func init() {
	cdc.EnableUSBCDC()
	machine.USBDev.Configure(machine.UARTConfig{})

}
