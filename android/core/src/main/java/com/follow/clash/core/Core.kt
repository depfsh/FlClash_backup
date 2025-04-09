package com.follow.clash.core

class Core {

    external fun stringFromJNI(): String

    companion object {
        init {
            System.loadLibrary("core")
        }
    }
}