//go:build android && cgo

package android_bride

/*
#include <stddef.h>
#include <stdint.h>
#include <malloc.h>

extern void (*mark_socket_func)(void *tun_interface, int fd);

extern int (*query_socket_uid_func)(void *tun_interface, int protocol, const char *source, const char *target);

void (*mark_socket_func)(void *tun_interface, int fd);

int (*query_socket_uid_func)(void *tun_interface, int protocol, const char *source, const char *target);

void mark_socket(void *interface, int fd) {
   mark_socket_func(interface, fd);
}

int query_socket_uid(void *interface, int protocol, char *source, char *target) {
    int result = query_socket_uid_func(interface, protocol, source, target);
    free(source);
    free(target);
    return result;
}
*/
import "C"
import "unsafe"

func MarkSocket(callback unsafe.Pointer, fd int) {
	C.mark_socket(callback, C.int(fd))
}

func QuerySocketUid(callback unsafe.Pointer, protocol int, source, target string) int {
	return int(C.query_socket_uid(callback, C.int(protocol), C.CString(source), C.CString(target)))
}
