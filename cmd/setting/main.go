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
		fmt.Println("Gains:", g)
	case "SetGains":
		client.SetGains(params)
	case "Settings":
		sets := client.Settings()
		s := settings.Settings{}.Merge(sets)
		fmt.Println("Settings:", s)
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
	default:
		fmt.Printf("Unknown sub-command: %s\n", sub)
		return
	}
	if err != nil {
		log.Fatal(err)
	}
}
