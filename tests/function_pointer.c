
int x;

void foo(void) {
    __atomic_store_n(&x, 1, __ATOMIC_SEQ_CST);
}

int main() {
    void (*bar)(void) = foo;

    bar();
    int y = __atomic_load_n(&x, __ATOMIC_SEQ_CST);

    return !(y == 1);
}
