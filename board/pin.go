package board

import (
	"machine"
)

const (
	CAN_CS   = machine.GPIO21
	CAN_RX   = machine.GPIO20
	CAN_TX   = machine.GPIO19
	CAN_SCK  = machine.GPIO18
	CAN_RST  = machine.GPIO17
	CAN_INT  = machine.GPIO16
	LCD_RS   = machine.GPIO8
	LCD_CS   = machine.GPIO9
	SPI1_SCK = machine.GPIO10
	SPI1_TX  = machine.GPIO11
	SPI1_RX  = machine.GPIO12
	SD_CS    = machine.GPIO13
	LED1     = machine.GPIO14
	LED2     = machine.GPIO15
	SW1      = machine.GPIO24
	SW2      = machine.GPIO23
	SW3      = machine.GPIO22
)

var (
	SLIDER1 machine.ADC
	SLIDER2 machine.ADC
)

func init() {
	machine.InitADC()
	SLIDER1 = machine.ADC{Pin: machine.ADC3}
	SLIDER2 = machine.ADC{Pin: machine.ADC2}
	SLIDER1.Configure(machine.ADCConfig{})
	SLIDER2.Configure(machine.ADCConfig{})
}
