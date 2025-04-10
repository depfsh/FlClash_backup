#include <jni.h>
#include <string>

#ifdef _LIBCLASH

#include "libclash.h"

extern "C"
JNIEXPORT void JNICALL
Java_com_follow_clash_core_Core_startTun(JNIEnv *env, jobject thiz, jint fd, jobject mark_socket,
                                         jobject query_socket_uid) {

    startTUN(fd);
}

extern "C"
JNIEXPORT void JNICALL
Java_com_follow_clash_core_Core_stopTun(JNIEnv *env, jobject thiz) {
    stopTun();
}
#endif
