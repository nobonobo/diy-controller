package main

import (
	"runtime"
	"time"

	"machine/usb"
	"machine/usb/hid/joystick"

	"github.com/marben/irpc"

	"github.com/nobonobo/diy-controller/board"
	"github.com/nobonobo/diy-controller/controller"
	"github.com/nobonobo/diy-controller/effects"
	"github.com/nobonobo/diy-controller/motor"
	"github.com/nobonobo/diy-controller/pid"
	"github.com/nobonobo/diy-controller/service"
	"github.com/nobonobo/diy-controller/stdio"
)

const MaxUserEffects = 8

var (
	pool = effects.NewEffectPool(MaxUserEffects)
	ph   = pid.NewPIDHandler(pool)
	js   = joystick.UseSettings(pid.Definitions,
		ph.RxHandler, ph.SetupHandler, pid.Descriptor)
)

func run(cntl *controller.Controller) {
	impl := &Service{controller: cntl}
	svc := service.NewServiceIrpcService(impl)
	if err := impl.Load(); err != nil {
		impl.Reset()
		//impl.Store()
	}
	conn := stdio.NewStdio()
	defer conn.Close()
	ep := irpc.NewEndpoint(conn, irpc.WithEndpointServices(svc))
	defer ep.Close()
	<-ep.Context().Done()
}

func init() {
	//usb.VendorID = 0x2341
	//usb.ProductID = 0x8036
	usb.Product = "DIY Steering Controller"
	usb.Manufacturer = "Switch Science"
	//cdc.EnableUSBCDC()
	//machine.USBDev.Configure(machine.UARTConfig{})
	board.LCD.Show(board.Logo)
	board.LCD.Display()
}

func main() {
	/*
		for !machine.Serial.DTR() {
			time.Sleep(100 * time.Millisecond)
		}
	*/
	can, err := board.NewCan(motor.CanRate, motor.CanExtended)
	if err != nil {
		//println(err)
		select {}
	}
	cntl := controller.New(pool)
	cntl.SetSettings(motor.DefaultSettings())
	mot := motor.New(can)
	if err := mot.Setup(); err != nil {
		//println(err)
		select {}
	}
	tick := time.NewTicker(time.Millisecond)
	input := new(controller.Input)
	cnt := 0
	//println("setup completed")
	go func() {
		for {
			run(cntl)
			runtime.GC()
		}
	}()
	for range tick.C {
		cnt++
		/*
			if (cnt/500)%2 == 0 {
				board.LED2.Low()
			} else {
				board.LED2.High()
			}
		*/
		state, err := mot.State()
		if err != nil {
			//println(err)
			select {}
		}
		input.Angle = state.Angle
		input.Velocity = state.Velocity
		out := cntl.Update(input, 0)
		steering := int(int64(32767) * int64(out.Angle) / int64(cntl.Settings().HalfOfL2L))
		if steering > 32767 {
			steering = 32767
		} else if steering < -32767 {
			steering = -32767
		}
		if cnt%10 == 0 {
			js.SetAxis(0, steering)
			js.SetAxis(2, steering)
			gamepad := cntl.GamePad()
			if gamepad != nil {
				js.SetAxis(1, int(gamepad.YAxis))
				js.SetAxis(3, int(gamepad.RxAxis))
				js.SetAxis(4, int(gamepad.RyAxis))
				js.SetAxis(5, int(gamepad.RzAxis))
				js.Buttons[0] = byte(gamepad.Buttons & 0xff)
				js.Buttons[1] = byte(gamepad.Buttons >> 8 & 0xff)
				js.SetHat(0, joystick.HatDirection(gamepad.Hat))
			}
			js.SendState()
		}
		/*
			if cnt%1000 == 0 {
				fmt.Printf("steering: %d, out: %+v\n", steering, out)
			}
		*/
		if err := mot.Output(out.Power); err != nil {
			//println(err)
			select {}
		}
	}
}
