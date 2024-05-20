#include <windows.h>
#include <iostream>

int main() {
    while (true) {
		// Sleep(10);

		for (int key = 0; key <= 254; key++)
		{
			if (GetAsyncKeyState(key) == -32767) {
                std::cout << "pressed! v2 " << int(-32767) << std::endl;
			}
		}
	}

	return 0;
}
