#include <pthread.h>
#include <stdatomic.h>

#define N 3

atomic_int g_x;

void *run(void *unused)
{
    g_x++;
    return NULL;
}

int main(void)
{
    pthread_t threads[N];
    for (size_t i = 0; i < N; i++)
    {
        pthread_create(&threads[i], NULL, run, NULL);
    }
    for (size_t i = 0; i < N; i++)
    {
        pthread_join(threads[i], NULL);
    }
    return 0;
}
