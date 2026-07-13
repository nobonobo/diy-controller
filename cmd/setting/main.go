package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/marben/irpc"
	"github.com/nobonobo/q16"
	"go.bug.st/serial"

	"github.com/nobonobo/diy-controller/service"
	"github.com/nobonobo/diy-controller/settings"
)

type wrapper struct {
	serial.Port
}

func (w *wrapper) Read(p []byte) (int, error) {
	n, err := w.Port.Read(p)
	log.Printf("RD: %X/%d/%v", p[:n], n, err)
	return n, err
}

func (w *wrapper) Write(p []byte) (int, error) {
	n, err := w.Port.Write(p)
	log.Printf("WR: %X/%d/%v", p[:n], n, err)
	return n, err
}

func command(client *service.ServiceIrpcClient, sub string, params map[string]int32) error {
	switch sub {
	case "Gains":
		gains := client.Gains()
		g := settings.NewGains().Merge(gains)
		fmt.Printf("Gains: %+v\n", g)
	case "SetGains":
		client.SetGains(params)
	case "Settings":
		sets := client.Settings()
		s := settings.Settings{}.Merge(sets)
		fmt.Printf("Settings: %+v\n", s)
	case "SetSettings":
		client.SetSettings(params)
	case "Store":
		if err := client.Store(); err != nil {
			log.Fatal(err)
		}
	case "Load":
		if err := client.Load(); err != nil {
			log.Fatal(err)
		}
	case "Reset":
		if err := client.Reset(); err != nil {
			log.Fatal(err)
		}
	case "StopAll":
		if err := client.StopAll(); err != nil {
			log.Fatal(err)
		}
	case "StopVibration":
		idx, ok := params["Idx"]
		if !ok {
			idx = int32(q16.Zero)
		}
		idx = idx >> q16.ShiftBits
		if err := client.StopVibration(int(idx)); err != nil {
			log.Fatal(err)
		}
	case "StartVibration":
		idx, ok := params["Idx"]
		if !ok {
			idx = int32(q16.Zero)
		}
		idx = idx >> q16.ShiftBits
		if err := client.StartVibration(int(idx)); err != nil {
			log.Fatal(err)
		}
	case "SetVibration":
		idx, ok := params["Idx"]
		if !ok {
			idx = int32(q16.Zero)
		}
		et := params["EffectType"] >> q16.ShiftBits
		idx = idx >> q16.ShiftBits
		log.Printf("params: %+v", params)
		if err := client.SetVibration(int(idx), &service.Vibration{
			Gain:       params["Gain"],
			EffectType: uint8(et),
			Duration:   params["Duration"],
			Frequency:  params["Frequency"],
		}); err != nil {
			log.Fatal(err)
		}
	case "SetEnvelope":
		idx, ok := params["Idx"]
		if !ok {
			idx = int32(q16.Zero)
		}
		idx = idx >> q16.ShiftBits
		if err := client.SetEnvelope(int(idx), &service.Envelope{
			AttackLevel: params["AttackLevel"],
			FadeLevel:   params["FadeLevel"],
			AttackTime:  params["AttackTime"],
			FadeTime:    params["FadeTime"],
		}); err != nil {
			log.Fatal(err)
		}
	case "ShowVibration":
		idx, ok := params["Idx"]
		if !ok {
			idx = int32(q16.Zero)
		}
		idx = idx >> q16.ShiftBits
		str, err := client.ShowVibration(int(idx))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", str)
	default:
		return errors.New("Unknown sub-command: " + sub)
	}
	return nil
}

func main() {
	port := "COM3"
	repeat := -1
	after := time.Duration(100 * time.Millisecond)
	flag.IntVar(&repeat, "repeat", repeat, "Repeat count")
	flag.DurationVar(&after, "after", after, "After duration")
	flag.StringVar(&port, "port", port, "Serial port")
	flag.Parse()
	sub := ""
	if len(flag.Args()) > 0 {
		sub = flag.Args()[0]
	}
	if sub == "" {
		println("Sub-command is required.\n")
		flag.Usage()
		return
	}
	conn, err := serial.Open(port, &serial.Mode{
		BaudRate: 115200 * 8,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	})
	if err != nil {
		log.Fatal(err)
	}
	//conn.SetReadTimeout(time.Second)
	defer conn.Close()
	ep := irpc.NewEndpoint(&wrapper{Port: conn})
	defer ep.Close()
	client, err := service.NewServiceIrpcClient(ep)
	if err != nil {
		log.Fatal(err)
	}
	params := map[string]int32{}
	for _, v := range flag.Args()[1:] {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Invalid parameter format: %s\n", v)
			continue
		}
		val, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			fmt.Printf("Invalid parameter value: %s\n", parts[1])
			continue
		}
		params[parts[0]] = int32(val * float64(q16.Scale))
	}
	cnt := 0
	for {
		if repeat > 0 {
			log.Printf("Executing command %d/%d", cnt+1, repeat)
		}
		err = command(client, sub, params)
		if err != nil {
			log.Fatal(err)
		}
		if repeat == 0 || (repeat > 0 && cnt+1 < repeat) {
			cnt++
			time.Sleep(after)
			continue
		}
		break
	}
}
