package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/marben/irpc"
	"github.com/nobonobo/q16"
	"go.bug.st/serial"

	"github.com/nobonobo/diy-controller/service"
	"github.com/nobonobo/diy-controller/settings"
)

func main() {
	port := "COM3"
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
	defer conn.Close()
	ep := irpc.NewEndpoint(conn)
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
		params["Idx"] = params["Idx"] >> q16.ShiftBits
		if err := client.StopVibration(int(params["Idx"])); err != nil {
			log.Fatal(err)
		}
	case "StartVibration":
		params["Idx"] = params["Idx"] >> q16.ShiftBits
		if err := client.StartVibration(int(params["Idx"])); err != nil {
			log.Fatal(err)
		}
	case "SetVibration":
		params["Idx"] = params["Idx"] >> q16.ShiftBits
		params["EffectType"] = params["EffectType"] >> q16.ShiftBits
		log.Printf("Setting vibration %d: %+v\n", int(params["Idx"]), params)
		if err := client.SetVibration(int(params["Idx"]), &service.Vibration{
			Gain:       params["Gain"],
			EffectType: uint8(params["EffectType"]),
			Duration:   params["Duration"],
			Frequency:  params["Frequency"],
		}); err != nil {
			log.Fatal(err)
		}
	case "SetEnvelope":
		params["Idx"] = params["Idx"] >> q16.ShiftBits
		if err := client.SetEnvelope(int(params["Idx"]), &service.Envelope{
			AttackLevel: params["AttackLevel"],
			FadeLevel:   params["FadeLevel"],
			AttackTime:  params["AttackTime"],
			FadeTime:    params["FadeTime"],
		}); err != nil {
			log.Fatal(err)
		}
	case "ShowVibration":
		params["Idx"] = params["Idx"] >> q16.ShiftBits
		str, err := client.ShowVibration(int(params["Idx"]))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", str)
	default:
		fmt.Printf("Unknown sub-command: %s\n", sub)
		return
	}
	if err != nil {
		log.Fatal(err)
	}
}
