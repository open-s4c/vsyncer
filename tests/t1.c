#include <stdatomic.h>

atomic_int x;

int main2() { return atomic_load(&x); }
int yyp;
int main() {
    int y = main2();
    int z = main2();
    yyp = z;
    return atomic_load(&x);
}
