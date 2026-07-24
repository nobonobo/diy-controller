package main

import (
	"context"
	"log"
	"syscall"
	"time"
	"unsafe"

	"github.com/0xcafed00d/joystick"
	"github.com/lxn/win"
	"github.com/marben/irpc"
	"go.bug.st/serial"

	"github.com/nobonobo/diy-controller/service"
)

type HHOOK uintptr

const (
	WH_KEYBOARD_LL = 13

	WM_KEYDOWN    = 0x0100
	WM_KEYUP      = 0x0101
	WM_SYSKEYDOWN = 0x0104
	WM_SYSKEYUP   = 0x0105
)

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	setWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	callNextHookEx      = user32.NewProc("CallNextHookEx")
	unhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
)

type KeyEvent struct {
	VK   uint16
	Code uint16
	Down bool
}

func proc(ctx context.Context, ch chan KeyEvent) error {
	defer close(ch)
	var (
		hook     HHOOK
		callback uintptr
	)
	keyboardProc := func(
		nCode int,
		wParam uintptr,
		lParam uintptr,
	) uintptr {
		if nCode >= 0 {
			kbd := (*KBDLLHOOKSTRUCT)(
				unsafe.Pointer(lParam),
			)
			switch wParam {
			case WM_KEYDOWN, WM_SYSKEYDOWN:
				select {
				case <-ctx.Done():
					return 0
				case ch <- KeyEvent{
					VK:   uint16(kbd.VkCode),
					Code: uint16(kbd.ScanCode),
					Down: true,
				}:
				}
			case WM_KEYUP, WM_SYSKEYUP:
				select {
				case <-ctx.Done():
					return 0
				case ch <- KeyEvent{
					VK:   uint16(kbd.VkCode),
					Code: uint16(kbd.ScanCode),
					Down: false,
				}:
				}
			}
		}
		ret, _, _ := callNextHookEx.Call(
			uintptr(hook),
			uintptr(nCode),
			wParam,
			lParam,
		)
		return ret
	}
	// GCされないよう保持
	callback = syscall.NewCallback(keyboardProc)
	ret, _, err := setWindowsHookExW.Call(
		WH_KEYBOARD_LL,
		callback,
		0,
		0,
	)
	hook = HHOOK(ret)
	if hook == 0 {
		return err
	}
	defer func() {
		unhookWindowsHookEx.Call(
			uintptr(hook),
		)
	}()
	// Windows message loop
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) != 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
	return nil
}

var hatTable = map[[4]bool]uint8{
	// up, down, left, right
	{true, false, false, false}:  0,
	{true, false, true, true}:    0,
	{true, false, false, true}:   1,
	{false, false, false, true}:  2,
	{true, true, false, true}:    2,
	{false, true, false, true}:   3,
	{false, true, false, false}:  4,
	{false, true, true, true}:    4,
	{false, true, true, false}:   5,
	{false, false, true, false}:  6,
	{true, true, true, false}:    6,
	{true, false, true, false}:   7,
	{false, false, false, false}: 8,
	{true, true, true, true}:     8,
	{true, true, false, false}:   8,
	{false, false, true, true}:   8,
}

func makeBits(b ...bool) uint32 {
	var bits uint32
	for i, v := range b {
		if v {
			bits |= 1 << uint(i)
		}
	}
	return bits
}

func main() {
	var js joystick.Joystick
	arrowKeys := [4]bool{}
	buttons := [8]bool{}
	for i := range 4 {
		j, err := joystick.Open(i)
		if err != nil {
			log.Print(err)
			continue
		}
		if j.AxisCount() == 6 && j.ButtonCount() == 8 {
			js = j
			break
		}
	}
	if js == nil {
		log.Fatalln("not found controller")
	}
	conn, err := serial.Open("COM3", &serial.Mode{
		BaudRate: 115200 * 8,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	})
	if err != nil {
		log.Fatalln(err)
	}
	ep := irpc.NewEndpoint(conn)
	defer ep.Close()
	svc, err := service.NewServiceIrpcClient(ep)
	if err != nil {
		log.Fatalln(err)
	}
	tick := time.NewTicker(10 * time.Millisecond)
	events := make(chan KeyEvent, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := proc(ctx, events); err != nil {
			log.Fatal(err)
		}
	}()
	for range tick.C {
		select {
		case m, ok := <-events:
			if !ok {
				break
			}
			//log.Println("ev:", m)
			switch m.VK {
			case win.VK_SPACE, win.VK_RETURN, 'Z': // A
				buttons[0] = m.Down
			case win.VK_ESCAPE, win.VK_BACK, 'X': // B
				buttons[1] = m.Down
			case win.VK_SHIFT, 'C': // X
				buttons[2] = m.Down
			case win.VK_CONTROL, 'V': // Y
				buttons[3] = m.Down
			case 'Q': // LB
				buttons[4] = m.Down
			case 'E': // RB
				buttons[5] = m.Down
			case '1': // LT
				buttons[6] = m.Down
			case '3': // RT
				buttons[7] = m.Down
			case win.VK_UP, 'W':
				arrowKeys[0] = m.Down
				if arrowKeys[1] {
					arrowKeys[1] = false
				}
			case win.VK_DOWN, 'S':
				if arrowKeys[0] {
					arrowKeys[0] = false
				}
				arrowKeys[1] = m.Down
			case win.VK_LEFT, 'A':
				arrowKeys[2] = m.Down
				if arrowKeys[3] {
					arrowKeys[3] = false
				}
			case win.VK_RIGHT, 'D':
				if arrowKeys[2] {
					arrowKeys[2] = false
				}
				arrowKeys[3] = m.Down
			}
		case <-tick.C:
			state, _ := js.Read()
			//fmt.Println(state.AxisData, state.Buttons)
			if err := svc.Send(&service.GamePad{
				YAxis:   int16(state.AxisData[0]),
				RxAxis:  int16(state.AxisData[1]),
				RyAxis:  int16(state.AxisData[2]),
				RzAxis:  int16(state.AxisData[3]),
				Buttons: uint16(state.Buttons<<8) | uint16(makeBits(buttons[:]...)&0xff),
				Hat:     hatTable[arrowKeys],
			}); err != nil {
				log.Print(err)
			}
		}
	}
}
