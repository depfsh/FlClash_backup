#pragma once

#include <stdlib.h>

extern void (*mark_socket_func)(void *tun_interface, int fd);

extern int (*query_socket_uid_func)(void *tun_interface, int protocol, const char *source, const char *target);

extern void (*complete_func)(void *completable, const char *exception);

extern void (*fetch_report_func)(void *fetch_callback, const char *status_json);

extern void mark_socket(void *interface, int fd);

extern int query_socket_uid(void *interface, int protocol, char *source, char *target);