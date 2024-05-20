#include "counter.h"
#include "framework.h"

#include <iostream>
#include <Windows.h>

#include <signal.h>
#include <string>
#include <vector>
#include <thread>
#include <mutex>

HHOOK eHook = NULL;
HHOOK mHook = NULL;

const char *banner = "\n\
    qAPM Tracker %s \n\
\n";

const std::string version = "v1.2.1";

void tick() {
    while(true) {
        Sleep(1000);
        incrementSecond();
    }
}

LRESULT mouseProc(int nCode, WPARAM wparam, LPARAM lparam)
{
    if (nCode < 0)
        return CallNextHookEx(eHook, nCode, wparam, lparam);

    if (wparam == WM_LBUTTONDOWN ||
        wparam == WM_RBUTTONDOWN ||
        wparam == WM_XBUTTONDOWN ||
        wparam == WM_MBUTTONDOWN)
        addAction();

    return CallNextHookEx(eHook, nCode, wparam, lparam);
}


LRESULT keyboardProc(int nCode, WPARAM wparam, LPARAM lparam)
{
    if (nCode < 0)
        return CallNextHookEx(mHook, nCode, wparam, lparam);

    if (wparam == WM_KEYDOWN || wparam == WM_SYSKEYDOWN)
        addAction();

    return CallNextHookEx(mHook, nCode, wparam, lparam);
}

void signal_callback_handler(int signum) {
    std::cout << "Caught signal " << signum << std::endl;
    exit(signum);
}

static LRESULT CALLBACK wndProc(HWND hwnd, UINT message, WPARAM wParam, LPARAM lParam)
{
    switch (message)
    {
    case WM_PAINT:
    {
        PAINTSTRUCT paintStruct;
        HDC hdc = BeginPaint(hwnd, &paintStruct);

        RECT rect;
        GetClientRect(hwnd, &rect);

        std::string text = " : APM ";
        text = std::to_string(currentAPM()) + text;

        std::wstring widestr = std::wstring(text.begin(), text.end());
        const wchar_t* widecstr = widestr.c_str();

        DrawText(
            hdc,
            widecstr,
            -1,
            &rect,
            DT_RIGHT|DT_NOCLIP|DT_SINGLELINE|DT_VCENTER
        );

        EndPaint(hwnd, &paintStruct);
        break;

    }
    case WM_TIMER:
        RECT rect;
        GetClientRect(hwnd, &rect);
        InvalidateRect(hwnd, &rect, TRUE);
        break;
    case WM_CREATE:
        break;
    case WM_DESTROY:
        PostQuitMessage(0);
        break;
    default:
        return DefWindowProc(hwnd, message, wParam, lParam);
    }
    return 0;
}

int main()
{
    signal(SIGINT, signal_callback_handler);

    eHook = SetWindowsHookEx(WH_KEYBOARD_LL, (HOOKPROC)keyboardProc, GetModuleHandle(NULL), 0);
    mHook = SetWindowsHookEx(WH_MOUSE_LL, (HOOKPROC)mouseProc, GetModuleHandle(NULL), 0);

    HINSTANCE instance = GetModuleHandle(0);
    HCURSOR cursor = LoadCursor(0,IDC_ARROW);

    WNDCLASSEX wndclass = {
        sizeof(WNDCLASSEX),
        CS_HREDRAW | CS_VREDRAW, // style
        wndProc, // window proc
        0, // extra bytes following window class
        0, // extra bytes following window instance
        instance, // hInstance
        LoadIcon(0,IDI_APPLICATION), // hIcon
        cursor, // hCursor
        HBRUSH(COLOR_WINDOW + 1), // hbrBackground
        0, // MenuName
        TEXT("actions-per-minute-class"), // ClassName
        LoadIcon(0,IDI_APPLICATION) // small icon
    };

    bool isClassRegistered = RegisterClassEx(&wndclass);
    if (!isClassRegistered) {
        std::cout << "class not registered" << std::endl;
        exit(1);
    }

    int height = 25;
    int width = 70;
    int extraStyles = WS_EX_COMPOSITED | WS_EX_LAYERED | WS_EX_NOACTIVATE | WS_EX_TOPMOST | WS_EX_TRANSPARENT;
	int styles = WS_VISIBLE | WS_POPUP;
    HWND hwnd = CreateWindowEx(
        extraStyles,
        TEXT("actions-per-minute-class"),
        TEXT("actions-per-minute"),
        styles,
        GetSystemMetrics(SM_CXSCREEN)-width, // x
        height*3, // y
        width, // width
        height, // height,
        0, // parent
        0, // menu
        instance,
        NULL
    );

    std::thread t(tick);

    int timer = 500;
    SetTimer(hwnd, timer, timer, 0);

    printf(banner, version);

    MSG msg = { };
    while (WM_QUIT != msg.message)
    {
        while (PeekMessage(&msg, NULL, 0, 0, PM_REMOVE) > 0)
        {
            TranslateMessage(&msg);
            DispatchMessage(&msg);
        }
        Sleep(1);
    }

    bool keyboardUnhooked = UnhookWindowsHookEx(eHook);
    bool mouseUnhooked = UnhookWindowsHookEx(eHook);

    return EXIT_SUCCESS;
}
