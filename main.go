package main

import (
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
	js   = joystick.UseSettings(joystick.Definitions{
		ReportID:     1,
		ButtonCnt:    0,
		HatSwitchCnt: 0,
		AxisDefs: []joystick.Constraint{
			{MinIn: -32767, MaxIn: 32767, MinOut: -32767, MaxOut: 32767}, // X-Axis
			{MinIn: -32767, MaxIn: 32767, MinOut: -32767, MaxOut: 32767}, // Y-Axis
			{MinIn: -32767, MaxIn: 32767, MinOut: -32767, MaxOut: 32767},
			{MinIn: -32767, MaxIn: 32767, MinOut: -32767, MaxOut: 32767}, // Rx-Axis
			{MinIn: -32767, MaxIn: 32767, MinOut: -32767, MaxOut: 32767}, // Ry-Axis
		},
	}, ph.RxHandler, ph.SetupHandler, pid.Descriptor)
)

func run(cntl *controller.Controller) {
	defer recover()
	service := service.NewServiceIrpcService(&Service{controller: cntl})
	conn := stdio.NewStdio()
	defer conn.Close()
	ep := irpc.NewEndpoint(conn, irpc.WithEndpointServices(service))
	defer ep.Close()
	<-ep.Context().Done()
}

func init() {
	//usb.VendorID = 0x2341
	//usb.ProductID = 0x8036
	usb.Product = "DIY Steering Controller"
	usb.Manufacturer = "Switch Science"
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
		println(err)
		select {}
	}
	cntl := controller.New(pool)
	cntl.SetSettings(motor.DefaultSettings())
	mot := motor.New(can)
	if err := mot.Setup(); err != nil {
		println(err)
		select {}
	}
	tick := time.NewTicker(time.Millisecond)
	input := new(controller.Input)
	cnt := 0
	//println("setup completed")
	go func() {
		for {
			run(cntl)
		}
	}()
	for range tick.C {
		cnt++
		state, err := mot.State()
		if err != nil {
			println(err)
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
		js.SetAxis(0, steering)
		js.SetAxis(2, steering)
		js.SendState()
		/*
			if cnt%1000 == 0 {
				fmt.Printf("steering: %d, out: %+v\n", steering, out)
			}
		*/
		if err := mot.Output(out.Power); err != nil {
			println(err)
			select {}
		}
	}
}
