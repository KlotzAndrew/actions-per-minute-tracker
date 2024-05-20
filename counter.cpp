#include <mutex>
#include <vector>
#include <iostream>

#include "counter.h"

#define windowSize 60

std::mutex mtx;
int actionsPerSecond [windowSize] = {0};
int rollingActionCount = 0;
int totalSeconds = 0;

void addAction()
{
    const std::lock_guard<std::mutex> lock(mtx);

    rollingActionCount++;
    int currentSecond = totalSeconds % windowSize;

    actionsPerSecond[currentSecond]++;
}

int currentAPM()
{
    if (totalSeconds == 0)
        return 0;

    if (totalSeconds < windowSize)
        return static_cast<int>(static_cast<float>(rollingActionCount) * (static_cast<float>(windowSize) / static_cast<float>(totalSeconds)));

    return rollingActionCount;
}

void incrementSecond() {
    const std::lock_guard<std::mutex> lock(mtx);

    totalSeconds++;

    int currentSecond = totalSeconds % windowSize;
    rollingActionCount -= actionsPerSecond[currentSecond];
    actionsPerSecond[currentSecond] = 0;

    std::cout << ".";
}
