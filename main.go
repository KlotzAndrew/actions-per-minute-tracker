package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage          = user32.NewProc("GetMessageW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
)

func main() {
	fmt.Println("starting main...")
	cb := callback{}
	hookKeyboard := SetWindowsHookEx(WH_KEYBOARD_LL, cb.keyboardCallback, 0, 0)
	hookMouse := SetWindowsHookEx(WH_MOUSE_LL, cb.mouseCallback, 0, 0)

	defer func() {
		UnhookWindowsHookEx(hookKeyboard)
		UnhookWindowsHookEx(hookMouse)
	}()

	cb.hookKeyboard = hookKeyboard
	cb.hookMouse = hookMouse

	var msg MSG
	for {
		fmt.Println("looping...")
		msgVal := GetMessage(&msg, 0, 0, 0)
		if msgVal <= 0 {
			fmt.Println("bad msg val", msgVal)
			break
		}
		fmt.Println("got msg", msg.WParam)
	}
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowshookexw
const (
	WH_KEYBOARD_LL = 13
	WH_MOUSE_LL    = 14
)

type callback struct {
	hookKeyboard HHOOK
	hookMouse    HHOOK
}

// https://docs.microsoft.com/en-us/windows/win32/winprog/windows-data-types
type (
	WPARAM  uintptr
	LPARAM  uintptr
	LRESULT uintptr
	HHOOK   uintptr
)

func (c *callback) keyboardCallback(code int, wparam WPARAM, lparam LPARAM) LRESULT {
	if code >= 0 {
		if wparam == WM_KEYDOWN {
			fmt.Println("got called keyboard", wparam)
		}
	}
	return CallNextHookEx(c.hookKeyboard, code, wparam, lparam)
}

func (c *callback) mouseCallback(code int, wparam WPARAM, lparam LPARAM) LRESULT {
	if code >= 0 {
		if wparam == WM_LBUTTONDOWN ||
			wparam == WM_RBUTTONDOWN ||
			wparam == WM_XBUTTONDOWN ||
			wparam == WM_MBUTTONDOWN {
			fmt.Println("got called mouse", wparam)
		} else {
		}
	}
	return CallNextHookEx(c.hookMouse, code, wparam, lparam)
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowshookexw
func SetWindowsHookEx(idHook int, lpfn HOOKPROC, hMod uintptr, dwThreadId uintptr) HHOOK {
	ret, _, _ := procSetWindowsHookEx.Call(
		uintptr(idHook),
		uintptr(syscall.NewCallback(lpfn)),
		uintptr(hMod),
		uintptr(dwThreadId),
	)
	return HHOOK(ret)
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-unhookwindowshookex
func UnhookWindowsHookEx(hhk HHOOK) bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(
		uintptr(hhk),
	)
	return ret != 0
}

// https://docs.microsoft.com/en-us/windows/win32/winmsg/lowlevelkeyboardproc
type HOOKPROC func(int, WPARAM, LPARAM) LRESULT

// http://msdn.microsoft.com/en-us/library/windows/desktop/ms644958.aspx
type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

// http://msdn.microsoft.com/en-us/library/windows/desktop/dd162805.aspx
type POINT struct {
	X, Y int32
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getmessage
func GetMessage(msg *MSG, hwnd uintptr, msgFilterMin, msgFilterMax uint32) int {
	ret, _, _ := procGetMessage.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	return int(ret)
}

// https://docs.microsoft.com/en-us/windows/win32/winmsg/lowlevelkeyboardproc
func CallNextHookEx(hhk HHOOK, nCode int, wparam WPARAM, lparam LPARAM) LRESULT {
	ret, _, _ := procCallNextHookEx.Call(
		uintptr(hhk),
		uintptr(nCode),
		uintptr(wparam),
		uintptr(lparam),
	)
	return LRESULT(ret)
}

// https://docs.microsoft.com/en-us/windows/win32/inputdev/keyboard-input-notifications
const (
	WM_KEYDOWN = 256
)

// https://docs.microsoft.com/en-us/windows/win32/inputdev/mouse-input-notifications
const (
	WM_LBUTTONDOWN = 513
	WM_RBUTTONDOWN = 516
	WM_MBUTTONDOWN = 519
	WM_XBUTTONDOWN = 523
)
