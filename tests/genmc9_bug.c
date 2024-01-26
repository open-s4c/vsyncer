#include <stdatomic.h>
#include <stdbool.h>
#include <pthread.h>
#include <stdio.h>


struct student {
    int age;
    char *name;
};

int main(void)
{
    struct student a = {0}, b = {0};
    a.age = 10;
    a.name = "genmc-10";
    b = a;
    printf("%d %s \n", b.age, b.name);
    return 0;
}
