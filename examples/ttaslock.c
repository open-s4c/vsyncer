#include <assert.h>
#include <pthread.h>
#include <stdatomic.h>
#include "../include/await_while.h"

#define NTHREADS 2

atomic_int l;

void
lock()
{
    do {
        await_while (l == 1)
            ;
    } while (atomic_exchange(&l, 1) == 1);
}

void
unlock()
{
    l = 0;
}


int x, y;

void *
run(void *_)
{
    lock();
    x++;
    y++;
    unlock();
    return 0;
}

int
main()
{
    pthread_t t[NTHREADS];

    for (int i = 0; i < NTHREADS; i++)
        pthread_create(&t[i], 0, run, 0);

    for (int i = 0; i < NTHREADS; i++)
        pthread_join(t[i], 0);

    assert(x == y);
    assert(x == NTHREADS);
    return 0;
}
