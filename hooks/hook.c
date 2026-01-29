#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <dlfcn.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <string.h>
#include <errno.h>

// Function pointer for original execve
typedef int (*execve_fn)(const char *pathname, char *const argv[], char *const envp[]);
static execve_fn real_execve = NULL;

void __attribute__((constructor)) init(void) {
    real_execve = (execve_fn)dlsym(RTLD_NEXT, "execve");
}

static void send_notification(const char *pathname, char *const argv[]) {
    const char *sock_path = getenv("AMUX_HOOK_SOCKET");
    if (!sock_path) return;

    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    if (fd < 0) return;

    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, sock_path, sizeof(addr.sun_path) - 1);

    if (connect(fd, (struct sockaddr*)&addr, sizeof(addr)) == 0) {
        // Construct simple JSON payload
        // Note: In production C code, handle buffer limits safely.
        char buf[4096];
        int pid = getpid();
        int ppid = getppid();
        
        // Count args to avoid buffer overflow roughly
        // ...
        
        int n = snprintf(buf, sizeof(buf), 
            "{\"pid\":%d,\"ppid\":%d,\"cmd\":\"%s\"}\n", 
            pid, ppid, pathname);
            
        if (n > 0) {
            write(fd, buf, n);
        }
    }
    close(fd);
}

int execve(const char *pathname, char *const argv[], char *const envp[]) {
    if (!real_execve) {
        real_execve = (execve_fn)dlsym(RTLD_NEXT, "execve");
    }

    send_notification(pathname, argv);

    return real_execve(pathname, argv, envp);
}

