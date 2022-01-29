package main

import (
	"fmt"
	"log"
	"time"

	"actions-per-minute-tracker/win32"
)

const refreshSignal = 35001

type APMTracker struct {
	hookKeyboard win32.HHOOK
	hookMouse    win32.HHOOK

	actionsPerSecond   []uint16
	rollingActionCount uint
	done               chan (bool)
	newActions         chan (int)
}

func newAPMTracker() *APMTracker {
	tracker := &APMTracker{
		done:             make(chan bool),
		actionsPerSecond: []uint16{0},
		newActions:       make(chan int),
	}
	tracker.Start()

	return tracker
}

const windowSize = 60

func (c *APMTracker) Start() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-c.done:
				return
			case <-c.newActions:
				currentSecond := len(c.actionsPerSecond) - 1
				c.actionsPerSecond[currentSecond]++
			case <-ticker.C:
				currentSecond := len(c.actionsPerSecond) - 1
				c.rollingActionCount += uint(c.actionsPerSecond[currentSecond])
				if currentSecond >= windowSize {
					c.rollingActionCount -= uint(c.actionsPerSecond[currentSecond-windowSize])
				}
				c.actionsPerSecond = append(c.actionsPerSecond, 0)
			}
		}
	}()
}

func (c *APMTracker) currentAPM() uint {
	switch currentWindowSize := len(c.actionsPerSecond) - 1; {
	case currentWindowSize >= windowSize:
		return c.rollingActionCount
	case currentWindowSize == 0:
		return 0
	default:
		return adjustFirstMinute(c.rollingActionCount, len(c.actionsPerSecond)-1)
	}
}

func adjustFirstMinute(rollingActions uint, currentWindowSize int) uint {
	multiplier := float64(windowSize) / float64(currentWindowSize)
	val := float64(rollingActions) * multiplier
	return uint(val)
}

func (c *APMTracker) addAction() {
	c.newActions <- 1
}

func (c *APMTracker) keyboardCallback(code int, wparam win32.WPARAM, lparam win32.LPARAM) win32.LRESULT {
	if code >= 0 {
		if wparam == win32.WM_KEYDOWN || wparam == win32.WM_SYSKEYDOWN {
			c.addAction()
		} else if wparam == win32.WM_KEYUP || wparam == win32.WM_SYSKEYUP {
			// ignored
		} else {
			fmt.Println("missed kboard cb", wparam, lparam)
		}
	}
	return win32.CallNextHookEx(c.hookKeyboard, code, wparam, lparam)
}

func (c *APMTracker) mouseCallback(code int, wparam win32.WPARAM, lparam win32.LPARAM) win32.LRESULT {
	if code >= 0 {
		if wparam == win32.WM_LBUTTONDOWN ||
			wparam == win32.WM_RBUTTONDOWN ||
			wparam == win32.WM_XBUTTONDOWN ||
			wparam == win32.WM_MBUTTONDOWN {
			c.addAction()
		} else {
			// no need to capture these
		}
	}
	return win32.CallNextHookEx(c.hookMouse, code, wparam, lparam)
}

func (r *APMTracker) windowProc(hwnd win32.HWND, msg uint32, wparam win32.WPARAM, lparam win32.LPARAM) win32.LRESULT {
	var paintStruct win32.PAINTSTRUCT

	switch msg {
	case win32.WM_PAINT:
		hdc := win32.BeginPaint(hwnd, &paintStruct)

		var rect win32.RECT
		win32.GetClientRect(hwnd, &rect)

		text := fmt.Sprintf("%d APM ", r.currentAPM())
		win32.DrawText(
			hdc,
			text,
			rect,
			win32.DT_RIGHT|win32.DT_NOCLIP|win32.DT_SINGLELINE|win32.DT_VCENTER,
		)
		win32.EndPaint(hwnd, &paintStruct)

		return 0
	case win32.WM_MOUSEMOVE, win32.WM_NCHITTEST, win32.WM_NCMOUSEMOVE, win32.WM_GETICON, win32.WM_LBUTTONDOWN, win32.WM_LBUTTONUP:
		ret := win32.DefWindowProc(hwnd, msg, wparam, lparam)
		return ret
	case win32.WM_SETCURSOR:
		ret := win32.DefWindowProc(hwnd, msg, wparam, lparam)
		return ret
	case refreshSignal:
		var rect win32.RECT
		win32.GetClientRect(hwnd, &rect)
		win32.InvalidateRect(hwnd, &rect)
		return 0
	case win32.WM_CLOSE:
		win32.DestroyWindow(hwnd)
	case win32.WM_DESTROY:
		win32.PostQuitMessage(0)
	default:
		ret := win32.DefWindowProc(hwnd, msg, wparam, lparam)
		return ret
	}
	return 0
}

func main() {
	fmt.Println("starting main...")
	// setup apm tracker
	tracker := newAPMTracker()
	hookKeyboard := win32.SetWindowsHookEx(win32.WH_KEYBOARD_LL, tracker.keyboardCallback, 0, 0)
	hookMouse := win32.SetWindowsHookEx(win32.WH_MOUSE_LL, tracker.mouseCallback, 0, 0)

	defer func() {
		win32.UnhookWindowsHookEx(hookKeyboard)
		win32.UnhookWindowsHookEx(hookMouse)
	}()

	tracker.hookKeyboard = hookKeyboard
	tracker.hookMouse = hookMouse

	// setup window
	className := "apm-window-object"

	instance, err := win32.GetModuleHandle()
	if err != nil {
		log.Println(err)
		return
	}

	cursor, err := win32.LoadCursorResource(win32.IDC_ARROW)
	if err != nil {
		log.Println(err)
		return
	}

	wndClass := win32.NewWNDClasss(className, tracker.windowProc, instance, cursor)
	if _, err = win32.RegisterClassEx(&wndClass); err != nil {
		log.Println(err)
		return
	}

	height := 25
	width := 70
	var extraStyles uint32 = win32.WS_EX_COMPOSITED | win32.WS_EX_LAYERED | win32.WS_EX_NOACTIVATE | win32.WS_EX_TOPMOST | win32.WS_EX_TRANSPARENT
	var styles uint32 = win32.WS_VISIBLE | win32.WS_POPUP
	hwnd := win32.CreateWindow(
		extraStyles, // extra style
		className,
		"Actions Per Minute Tracker", // name
		uint32(styles),               // style
		int64(win32.GetSystemMetrics(win32.SM_CXSCREEN))-int64(width), // x
		int64(height*3), // y
		int64(width),    // width
		int64(height),   // height
		0,               // parent
		0,               // menu
		instance,
	)

	win32.SetWindowPos(
		hwnd,
		^win32.HWND(0),
		0,
		0,
		0,
		0,
		win32.SWP_NOACTIVATE|win32.SWP_NOMOVE|win32.SWP_NOSIZE|win32.SWP_SHOWWINDOW,
	)

	// refresh loop
	done := make(chan int)
	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				win32.SendMessage(hwnd, refreshSignal, 0, 0)
			}
		}
	}()

	// pull messages
	var msg win32.MSG
	for {
		msgVal := win32.GetMessage(&msg, 0, 0, 0)
		if msgVal <= 0 {
			fmt.Println("bad msg val", msgVal)
			break
		}
		win32.TranslateMessage(&msg)
		win32.DispatchMessage(&msg)
	}
}
