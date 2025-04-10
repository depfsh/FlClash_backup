#include <jni.h>
#include <string>

#ifdef _LIBCLASH
#include "libclash.h"
#endif

extern "C" JNIEXPORT jstring JNICALL
Java_com_follow_clash_core_Core_stringFromJNI(
        JNIEnv* env,
        jobject /* this */) {
    std::string hello = "Hello from C++";
    return env->NewStringUTF(hello.c_str());
}

