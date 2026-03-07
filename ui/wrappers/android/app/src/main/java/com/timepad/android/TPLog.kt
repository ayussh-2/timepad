package com.timepad.android

import android.util.Log
import java.util.ArrayDeque

/** Single logcat tag "TIMEPAD" for all app logs. Filter: adb logcat TIMEPAD:D *:S */
internal object TPLog {
    const val TAG = "TIMEPAD"

    private const val MAX = 500
    private val buffer = ArrayDeque<String>(MAX)

    @Synchronized
    private fun record(level: String, ctx: String, msg: String) {
        if (buffer.size >= MAX) buffer.pollFirst()
        buffer.addLast("$level [$ctx] $msg")
    }

    @Synchronized fun lines(): List<String> = buffer.toList()

    fun d(ctx: String, msg: String) {
        Log.d(TAG, "[$ctx] $msg")
        record("D", ctx, msg)
    }
    fun w(ctx: String, msg: String) {
        Log.w(TAG, "[$ctx] $msg")
        record("W", ctx, msg)
    }
    fun e(ctx: String, msg: String) {
        Log.e(TAG, "[$ctx] $msg")
        record("E", ctx, msg)
    }
    fun e(ctx: String, msg: String, t: Throwable) {
        Log.e(TAG, "[$ctx] $msg", t)
        record("E", ctx, "$msg — ${t.message}")
    }
}
