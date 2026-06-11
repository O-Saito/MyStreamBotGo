package keylistener

import (
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procSetWindowsHookEx = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx   = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage       = user32.NewProc("GetMessageW")
	procGetKeyState      = user32.NewProc("GetKeyState")
)

const (
	WH_KEYBOARD_LL = 13

	WM_KEYDOWN    = 0x0100
	WM_KEYUP      = 0x0101
	WM_SYSKEYDOWN = 0x0104
	WM_SYSKEYUP   = 0x0105
	WM_QUIT       = 0x0012
)

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PtX    int32
	PtY    int32
}

var (
	hookHandle   uintptr
	hookCallback uintptr
	keyHandler   func(vkCode uint32, flags uint32, wParam uintptr)
)

func startHook(handler func(vkCode uint32, flags uint32, wParam uintptr)) error {
	keyHandler = handler

	cb := syscall.NewCallback(func(nCode int, wParam, lParam uintptr) uintptr {
		if nCode >= 0 && keyHandler != nil {
			kbd := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
			keyHandler(kbd.VkCode, kbd.Flags, wParam)
		}
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	})

	hookCallback = cb

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		h, _, _ := procSetWindowsHookEx.Call(
			WH_KEYBOARD_LL,
			cb,
			0,
			0,
		)
		if h == 0 {
			return
		}
		hookHandle = h

		var msg MSG
		for {
			ret, _, _ := procGetMessage.Call(
				uintptr(unsafe.Pointer(&msg)),
				0,
				0,
				0,
			)
			if ret == 0 {
				break
			}
		}
	}()

	return nil
}

func stopHook() {
	if hookHandle != 0 {
		procUnhookWindowsHookEx.Call(hookHandle)
		hookHandle = 0
	}
	keyHandler = nil
}
