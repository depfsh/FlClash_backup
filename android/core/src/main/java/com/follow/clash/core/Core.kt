package com.follow.clash.core

import java.net.InetAddress
import java.net.InetSocketAddress
import java.net.URL

data object Core {
    private external fun startTun(
        fd: Int,
        cb: TunInterface
    )

    private fun parseInetSocketAddress(address: String): InetSocketAddress {
        val url = URL("https://$address")

        return InetSocketAddress(InetAddress.getByName(url.host), url.port)
    }

    fun startTun(
        fd: Int,
        markSocket: (Int) -> Boolean,
        querySocketUid: (protocol: Int, source: InetSocketAddress, target: InetSocketAddress) -> Int
    ) {
        startTun(fd, object : TunInterface {
            override fun markSocket(fd: Int) {
                markSocket(fd)
            }

            override fun querySocketUid(protocol: Int, source: String, target: String): Int {
                return querySocketUid(
                    protocol,
                    parseInetSocketAddress(source),
                    parseInetSocketAddress(target),
                )
            }
        });
    }

    external fun stopTun()

    init {
        System.loadLibrary("core")
    }
}