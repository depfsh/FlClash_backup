package com.follow.clash.core

import java.net.InetSocketAddress


class Core {

    external fun startTun(
        fd: Int,
        markSocket: (Int) -> Boolean,
        querySocketUid: (protocol: Int, source: InetSocketAddress, target: InetSocketAddress) -> Int
    )

    external fun stopTun()

    companion object {
        init {
            System.loadLibrary("core")
        }
    }
}