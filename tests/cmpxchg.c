int shared;
int expected;

int main()
{
    __atomic_compare_exchange_n(&shared, &expected, 1, 0, __ATOMIC_SEQ_CST,
__ATOMIC_SEQ_CST);
    return 0;
}
