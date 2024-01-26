#include <stdbool.h>
#include "../include/await_while.h"

bool x = true;

int main() {

    await_while (x) {}

    return 0;
}
