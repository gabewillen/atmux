#include <stdio.h>
#include <unistd.h>
#include <dlfcn.h>

// Placeholder for execve interception
// In real implementation, we'd use dlsym(RTLD_NEXT, "execve")
// and send data to the unix socket.

void __attribute__((constructor)) init(void) {
    // printf("Amux hook loaded\n");
}

