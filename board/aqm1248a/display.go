package aqm1248a

import (
	"image/color"
	"machine"
	"time"
)

var (
	dummy = make([]byte, 128)
)

/*
	SPI      = machine.SPI1
	LCD_RS   = machine.GPIO8
	LCD_CS   = machine.GPIO9
	SPI1_SCK = machine.GPIO10
	SPI1_TX  = machine.GPIO11
	SPI1_RX  = machine.GPIO12

	display := New(SPI, LCD_RS, LCD_CS, SPI1_SCK, SPI1_TX, SPI1_RX)
*/

// Display は6行x128列の1ビットモノクロディスプレイを扱うstruct
type Display struct {
	Image               [6][128]byte
	spi                 *machine.SPI
	rs, cs, sck, tx, rx machine.Pin
}

func (d *Display) cmd(cmd byte) {
	d.rs.Low()
	d.cs.Low()
	d.spi.Transfer(cmd)
	d.cs.High()
}

func (d *Display) data(b []byte) {
	d.rs.High()
	d.cs.Low()
	//dummy := make([]byte, len(b))
	d.spi.Tx(b, dummy)
	d.cs.High()
}

// NewDisplay はBlackとClear色を初期化して返す
func New(spi *machine.SPI, rs, cs, sck, tx, rx machine.Pin) *Display {
	d := &Display{spi: spi, rs: rs, cs: cs, sck: sck, tx: tx, rx: rx}
	d.rs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.cs.High()
	d.spi.Configure(machine.SPIConfig{
		Mode: 3,
		SCK:  d.sck,
		SDO:  d.tx,
		SDI:  d.rx,
	})
	d.cmd(0xae) // Display = OFF
	d.cmd(0xa0) // ADC  normal
	d.cmd(0xc8) // Reverse = OFF
	d.cmd(0xa3) // LCD bias = 1/7
	//内部レギュレータON
	d.cmd(0x2c)
	time.Sleep(time.Millisecond)
	d.cmd(0x2e)
	time.Sleep(time.Millisecond)
	d.cmd(0x2f)
	//コントラスト設定
	d.cmd(0x23) //Vo voltage regulator internal resistor ratio set
	d.cmd(0x81) //Electronic volume mode set
	d.cmd(0x1c) //Electronic volume register set
	//表示設定
	d.cmd(0xa4) //Display all point ON/OFF = normal
	d.cmd(0x40) //Display start line = 0
	d.cmd(0xa6) //Display normal/revers = normal
	d.cmd(0xaF) //Dsiplay = ON
	return d
}

// Size は表示サイズを返す（幅128ピクセル、高さ48ピクセル）
func (d *Display) Size() (width, height int16) {
	return 128, 48
}

// SetPixel はx,yに色をセットする
func (d *Display) SetPixel(x, y int16, c color.RGBA) {
	if x < 0 || x >= 128 || y < 0 || y >= 48 {
		return
	}
	line := y / 8
	bit := uint(1 << (y % 8))
	if c.A > 0x80 {
		d.Image[line][x] |= byte(bit)
	} else {
		d.Image[line][x] &^= byte(bit)
	}
}

func (d *Display) Clear() {
	for b := 0; b < 6; b++ {
		for i := 0; i < 128; i++ {
			d.Image[b][i] = 0
		}
	}
}

func (d *Display) Show(img [6][128]byte) {
	d.Image = img
}

// Display はドライバにメモリの内容を送るfunctionの例
func (d *Display) Display() error {
	for i := 0; i < 6; i++ {
		d.cmd(0xb0 + byte(i))
		d.cmd(0x10)
		d.cmd(0x00)
		d.data(d.Image[i][:])
	}
	return nil
}
