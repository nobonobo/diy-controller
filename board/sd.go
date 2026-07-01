package board

import "machine"

func init() {
	SD_CS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SD_CS.High()
}
