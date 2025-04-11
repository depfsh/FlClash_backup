#ifdef _LIBCLASH

#include <jni.h>
#include <string>
#include "jni_helper.h"
#include "libclash.h"

extern "C"
JNIEXPORT void JNICALL
Java_com_follow_clash_core_Core_startTun(JNIEnv *env, jobject thiz, jint fd, jobject cb) {
    auto interface = new_global(cb);
    startTUN(fd, interface);
}

extern "C"
JNIEXPORT void JNICALL
Java_com_follow_clash_core_Core_stopTun(JNIEnv *env, jobject thiz) {
    stopTun();
}


static jmethodID m_tun_interface_mark_socket;
static jmethodID m_tun_interface_query_socket_uid;

static void call_tun_interface_mark_socket_impl(void *tun_interface, int fd) {
    ATTACH_JNI();
    env->CallVoidMethod((jobject) tun_interface,
                        (jmethodID) m_tun_interface_mark_socket,
                        (jint) fd);
}

static int
call_tun_interface_query_socket_uid_impl(void *tun_interface, int protocol,
                                         const char *source,
                                         const char *target) {
    ATTACH_JNI();
    return env->CallIntMethod((jobject) tun_interface,
                              (jmethodID) m_tun_interface_query_socket_uid,
                              (jint) protocol,
                              (jstring) new_string(source),
                              (jstring) new_string(target));
}

extern "C"
JNIEXPORT jint JNICALL
JNI_OnLoad(JavaVM *vm, void *reserved) {
    JNIEnv *env = nullptr;
    if (vm->GetEnv((void **) &env, JNI_VERSION_1_6) != JNI_OK) {
        return JNI_ERR;
    }

    initialize_jni(vm, env);

    jclass c_tun_interface = find_class("com/follow/clash/core/TunInterface");

    m_tun_interface_mark_socket = find_method(c_tun_interface, "markSocket", "(I)V");
    m_tun_interface_query_socket_uid = find_method(c_tun_interface, "querySocketUid",
                                                   "(ILjava/lang/String;Ljava/lang/String;)I");

    registerCallbacks(&call_tun_interface_mark_socket_impl,
                      &call_tun_interface_query_socket_uid_impl);
    return JNI_VERSION_1_6;
}

#endif
