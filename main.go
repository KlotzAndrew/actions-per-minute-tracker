package main

import (
	"fmt"

	"actions-per-minute-tracker/win32"
)

type callback struct {
	hookKeyboard win32.HHOOK
	hookMouse    win32.HHOOK
}

func (c *callback) keyboardCallback(code int, wparam win32.WPARAM, lparam win32.LPARAM) win32.LRESULT {
	if code >= 0 {
		if wparam == win32.WM_KEYDOWN {
			fmt.Println("got called keyboard", wparam)
		}
	}
	return win32.CallNextHookEx(c.hookKeyboard, code, wparam, lparam)
}

func (c *callback) mouseCallback(code int, wparam win32.WPARAM, lparam win32.LPARAM) win32.LRESULT {
	if code >= 0 {
		if wparam == win32.WM_LBUTTONDOWN ||
			wparam == win32.WM_RBUTTONDOWN ||
			wparam == win32.WM_XBUTTONDOWN ||
			wparam == win32.WM_MBUTTONDOWN {
			fmt.Println("got called mouse", wparam)
		} else {
		}
	}
	return win32.CallNextHookEx(c.hookMouse, code, wparam, lparam)
}

func main() {
	fmt.Println("starting main...")
	cb := callback{}
	hookKeyboard := win32.SetWindowsHookEx(win32.WH_KEYBOARD_LL, cb.keyboardCallback, 0, 0)
	hookMouse := win32.SetWindowsHookEx(win32.WH_MOUSE_LL, cb.mouseCallback, 0, 0)

	defer func() {
		win32.UnhookWindowsHookEx(hookKeyboard)
		win32.UnhookWindowsHookEx(hookMouse)
	}()

	cb.hookKeyboard = hookKeyboard
	cb.hookMouse = hookMouse

	var msg win32.MSG
	for {
		fmt.Println("looping...")
		msgVal := win32.GetMessage(&msg, 0, 0, 0)
		if msgVal <= 0 {
			fmt.Println("bad msg val", msgVal)
			break
		}
		fmt.Println("got msg", msg.WParam)
	}
}
