package com.follow.clash.core

class Core {

    fun hello(){

    }

    external fun stringFromJNI(): String

    companion object {
        init {
            System.loadLibrary("core")
        }
    }
}