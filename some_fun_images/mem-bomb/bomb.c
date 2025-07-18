#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <string.h>

int main() {
    const size_t block_size = 10 * 1024 * 1024; // 10 MB
    size_t total_allocated = 0;
    while (1) {
        void* ptr = malloc(block_size);
        if (!ptr) {
            continue;
        }
        memset(ptr, 0, block_size); 
        total_allocated += block_size;
        sleep(2); 
    }
    sleep(1);
    return 0;
}
