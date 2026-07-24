package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
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

func proc(ch chan KeyEvent) error {
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
				ch <- KeyEvent{
					VK:   uint16(kbd.VkCode),
					Code: uint16(kbd.ScanCode),
					Down: true,
				}
			case WM_KEYUP, WM_SYSKEYUP:
				ch <- KeyEvent{
					VK:   uint16(kbd.VkCode),
					Code: uint16(kbd.ScanCode),
					Down: false,
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

func main() {
	ch := make(chan KeyEvent, 10)
	go func() {
		err := proc(ch)
		if err != nil {
			panic(err)
		}
	}()
	for event := range ch {
		fmt.Printf("event: %+v\n", event)
	}
}
