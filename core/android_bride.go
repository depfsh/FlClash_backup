//go:build android && cgo

package main

/*
#include <stdlib.h>

typedef void (*mark_socket_func)(void *tun_interface, int fd);

typedef int (*query_socket_uid_func)(void *tun_interface, int protocol, const char *source, const char *target);

static void call_mark_socket(mark_socket_func fn, void *tun_interface, int fd) {
    if (fn) {
        fn(tun_interface, fd);
    }
}

static int call_query_socket_uid(query_socket_uid_func fn, void *tun_interface, int protocol, const char *source, const char *target) {
    if (fn) {
        return fn(tun_interface, protocol, source, target);
    }
    return -1;
}
*/
import "C"
import (
	"unsafe"
)

var (
	globalCallbacks struct {
		markSocketFunc     C.mark_socket_func
		querySocketUidFunc C.query_socket_uid_func
	}
)

func markSocket(callback unsafe.Pointer, fd int) {
	if globalCallbacks.markSocketFunc != nil {
		C.call_mark_socket(globalCallbacks.markSocketFunc, callback, C.int(fd))
	}
}

func querySocketUid(callback unsafe.Pointer, protocol int, source, target string) int {
	if globalCallbacks.querySocketUidFunc == nil {
		return -1
	}
	s := C.CString(source)
	defer C.free(unsafe.Pointer(s))
	t := C.CString(target)
	defer C.free(unsafe.Pointer(t))
	return int(C.call_query_socket_uid(globalCallbacks.querySocketUidFunc, callback, C.int(protocol), s, t))
}

//export registerCallbacks
func registerCallbacks(MarkSocketFunc C.mark_socket_func, QuerySocketUidFunc C.query_socket_uid_func) {
	globalCallbacks.markSocketFunc = MarkSocketFunc
	globalCallbacks.querySocketUidFunc = QuerySocketUidFunc
}
