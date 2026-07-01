package board

import "machine"

func init() {
	SW1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	SW2.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	SW3.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	LED1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	LED2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	LED1.High()
	LED2.High()
}
